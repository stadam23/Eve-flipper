package sde

import (
	"testing"
)

// CalculateMaterialsWithME: baseQty = base * runs * (1 - ME/100), ceiling, at least runs.
// CalculateTimeWithTE: time * runs * (1 - TE/100).

func TestCalculateMaterialsWithME_Exact(t *testing.T) {
	bp := &Blueprint{
		Materials: []BlueprintMaterial{
			{TypeID: 34, Quantity: 1000},
			{TypeID: 35, Quantity: 500},
		},
	}
	// ME=0: 1.0 multiplier. 1000*5=5000, 500*5=2500
	got := bp.CalculateMaterialsWithME(5, 0)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Quantity != 5000 || got[1].Quantity != 2500 {
		t.Errorf("ME=0 runs=5: got %v, want [5000, 2500]", got)
	}

	// ME=10: 0.9 multiplier. 1000*5*0.9=4500, 500*5*0.9=2250
	got = bp.CalculateMaterialsWithME(5, 10)
	if got[0].Quantity != 4500 || got[1].Quantity != 2250 {
		t.Errorf("ME=10 runs=5: got %v, want [4500, 2250]", got)
	}

	// ME=10, 1 run: 1000*1*0.9=900 ceiling=900, 500*1*0.9=450. Min runs=1 so both >= 1 (ok)
	got = bp.CalculateMaterialsWithME(1, 10)
	if got[0].Quantity != 900 || got[1].Quantity != 450 {
		t.Errorf("ME=10 runs=1: got %v, want [900, 450]", got)
	}
}

func TestCalculateMaterialsWithME_ClampsME(t *testing.T) {
	bp := &Blueprint{Materials: []BlueprintMaterial{{TypeID: 1, Quantity: 100}}}
	// ME < 0 -> 0; ME > 10 -> 10
	got := bp.CalculateMaterialsWithME(1, -5)
	if got[0].Quantity != 100 {
		t.Errorf("ME=-5 should clamp to 0, got %v", got[0].Quantity)
	}
	got = bp.CalculateMaterialsWithME(1, 15)
	// ME 15 clamped to 10: 100*1*0.9 = 90
	if got[0].Quantity != 90 {
		t.Errorf("ME=15 clamped to 10: got %v, want 90", got[0].Quantity)
	}
}

func TestCalculateMaterialsWithME_FloorAtRuns(t *testing.T) {
	// Small quantity: 10 * 2 runs * 0.9 = 18. But floor is runs=2.
	bp := &Blueprint{Materials: []BlueprintMaterial{{TypeID: 1, Quantity: 10}}}
	got := bp.CalculateMaterialsWithME(2, 10)
	if got[0].Quantity != 18 {
		t.Errorf("10*2*0.9=18: got %v", got[0].Quantity)
	}
	// 1 * 5 runs * 0.99 (ME=1) = 4.95 -> ceiling 5. Min runs=5 so 5.
	bp2 := &Blueprint{Materials: []BlueprintMaterial{{TypeID: 1, Quantity: 1}}}
	got = bp2.CalculateMaterialsWithME(5, 1)
	if got[0].Quantity < 5 {
		t.Errorf("quantity must be >= runs: got %v", got[0].Quantity)
	}
}

func TestCalculateTimeWithTE_Exact(t *testing.T) {
	bp := &Blueprint{Time: 3600}
	// TE=0: 3600 * 1 * 1 = 3600
	if got := bp.CalculateTimeWithTE(1, 0); got != 3600 {
		t.Errorf("Time 3600 runs 1 TE 0 = %v, want 3600", got)
	}
	// TE=20: 3600 * 2 * 0.8 = 5760
	if got := bp.CalculateTimeWithTE(2, 20); got != 5760 {
		t.Errorf("Time 3600 runs 2 TE 20 = %v, want 5760", got)
	}
	// TE=10: 3600 * 3 * 0.9 = 9720
	if got := bp.CalculateTimeWithTE(3, 10); got != 9720 {
		t.Errorf("Time 3600 runs 3 TE 10 = %v, want 9720", got)
	}
}

func TestCalculateTimeWithTE_ClampsTE(t *testing.T) {
	bp := &Blueprint{Time: 1000}
	if got := bp.CalculateTimeWithTE(1, -1); got != 1000 {
		t.Errorf("TE=-1 should clamp to 0: got %v", got)
	}
	if got := bp.CalculateTimeWithTE(1, 25); got != 800 {
		t.Errorf("TE=25 clamp to 20: 1000*0.8=800, got %v", got)
	}
}

func TestGetBlueprintForProduct(t *testing.T) {
	ind := NewIndustryData()
	bp := &Blueprint{ProductTypeID: 645, ProductQuantity: 1}
	ind.Blueprints[123] = bp
	ind.ProductToBlueprint[645] = 123

	got, ok := ind.GetBlueprintForProduct(645)
	if !ok || got != bp {
		t.Errorf("GetBlueprintForProduct(645) = %v, %v; want bp, true", got, ok)
	}
	_, ok = ind.GetBlueprintForProduct(999)
	if ok {
		t.Error("GetBlueprintForProduct(999) should be false")
	}
}
