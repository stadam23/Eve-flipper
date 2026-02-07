package engine

import (
	"math"
	"testing"
	"time"

	"eve-flipper/internal/esi"
)

func TestVariance(t *testing.T) {
	// Sample variance (Bessel's correction): [1,2,3,4,5] mean=3, var = (4+1+0+1+4)/(5-1) = 2.5
	x := []float64{1, 2, 3, 4, 5}
	got := variance(x)
	if math.Abs(got-2.5) > 1e-9 {
		t.Errorf("variance(%v) = %v, want 2.5", x, got)
	}
	if variance(nil) != 0 {
		t.Errorf("variance(nil) want 0")
	}
	if variance([]float64{7}) != 0 {
		t.Errorf("variance(single) want 0")
	}
}

func TestCalibrateImpact(t *testing.T) {
	// Use dates within last 30 days so filterLastNDays keeps them
	base := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	d1 := time.Now().AddDate(0, 0, -9).Format("2006-01-02")
	d2 := time.Now().AddDate(0, 0, -8).Format("2006-01-02")
	d3 := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	d4 := time.Now().AddDate(0, 0, -6).Format("2006-01-02")
	history := []esi.HistoryEntry{
		{Date: base, Average: 100, Highest: 102, Lowest: 98, Volume: 1000, OrderCount: 10},
		{Date: d1, Average: 100, Highest: 103, Lowest: 97, Volume: 2000, OrderCount: 12},
		{Date: d2, Average: 101, Highest: 104, Lowest: 98, Volume: 1500, OrderCount: 11},
		{Date: d3, Average: 99, Highest: 102, Lowest: 96, Volume: 1200, OrderCount: 9},
		{Date: d4, Average: 100, Highest: 105, Lowest: 95, Volume: 1800, OrderCount: 14},
	}
	params := CalibrateImpact(history, 30)
	if !params.Valid {
		t.Fatal("expected Valid=true with 5 days of history")
	}
	if params.Amihud <= 0 {
		t.Errorf("expected positive Amihud, got %v", params.Amihud)
	}
	if params.Sigma <= 0 {
		t.Errorf("expected positive Sigma, got %v", params.Sigma)
	}
	if params.AvgDailyVolume <= 0 {
		t.Errorf("expected positive AvgDailyVolume, got %v", params.AvgDailyVolume)
	}
	if params.DaysUsed != 5 {
		t.Errorf("DaysUsed want 5, got %d", params.DaysUsed)
	}
}

func TestCalibrateImpact_TooFewDays(t *testing.T) {
	d1 := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	d2 := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	history := []esi.HistoryEntry{
		{Date: d1, Average: 100, Volume: 1000},
		{Date: d2, Average: 101, Volume: 1000},
	}
	params := CalibrateImpact(history, 30)
	if params.Valid {
		t.Error("expected Valid=false with only 2 days")
	}
}

func TestImpactLinearPct(t *testing.T) {
	// Amihud = 0.0001 (fractional per unit), Q = 1000
	// ΔP% = 0.0001 × 1000 × 100 = 10%
	got := ImpactLinearPct(0.0001, 1000)
	want := 10.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("ImpactLinearPct(0.0001, 1000) = %v, want %v", got, want)
	}
	if ImpactLinearPct(0.0001, 0) != 0 {
		t.Error("expected 0 for zero quantity")
	}
}

func TestImpactSqrtPct(t *testing.T) {
	// σ = 0.02, Q = 10000, V_daily = 10000
	// ΔP% = 0.02 × √(10000/10000) × 100 = 0.02 × 1 × 100 = 2%
	got := ImpactSqrtPct(0.02, 10000, 10000)
	want := 2.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("ImpactSqrtPct(0.02, 10000, 10000) = %v, want %v", got, want)
	}
	// Half the volume → impact × √0.5 ≈ 1.414%
	got2 := ImpactSqrtPct(0.02, 5000, 10000)
	want2 := 0.02 * math.Sqrt(0.5) * 100
	if math.Abs(got2-want2) > 1e-9 {
		t.Errorf("ImpactSqrtPct(0.02, 5000, 10000) = %v, want %v", got2, want2)
	}
	if ImpactSqrtPct(0.02, 0, 10000) != 0 {
		t.Error("expected 0 for zero quantity")
	}
	if ImpactSqrtPct(0.02, 1000, 0) != 0 {
		t.Error("expected 0 for zero daily volume")
	}
}

func TestOptimalSlicesVolume(t *testing.T) {
	// Q = 1000, V_daily = 10000, targetPct = 0.05
	// sliceSize = 500, n = ceil(1000/500) = 2
	n := OptimalSlicesVolume(1000, 10000, 0.05)
	if n != 2 {
		t.Errorf("OptimalSlicesVolume(1000, 10000, 0.05) = %d, want 2", n)
	}
	// Q = 100 → 1 slice (100 < 500)
	n2 := OptimalSlicesVolume(100, 10000, 0.05)
	if n2 != 1 {
		t.Errorf("OptimalSlicesVolume(100, 10000, 0.05) = %d, want 1", n2)
	}
	// Edge: zero daily volume → 1
	if OptimalSlicesVolume(1000, 0, 0.05) != 1 {
		t.Error("expected 1 for zero daily volume")
	}
}

func TestEstimateImpact(t *testing.T) {
	params := ImpactParams{
		Amihud:         0.00001,
		Sigma:          0.02,
		SigmaSq:        0.0004,
		AvgDailyVolume: 10000,
		Valid:          true,
	}
	est := EstimateImpact(params, 5000, 100.0)
	if est.LinearImpactPct <= 0 {
		t.Errorf("expected positive LinearImpactPct, got %v", est.LinearImpactPct)
	}
	if est.SqrtImpactPct <= 0 {
		t.Errorf("expected positive SqrtImpactPct, got %v", est.SqrtImpactPct)
	}
	if est.RecommendedImpactPct <= 0 {
		t.Errorf("expected positive RecommendedImpactPct, got %v", est.RecommendedImpactPct)
	}
	if est.RecommendedImpactISK <= 0 {
		t.Errorf("expected positive RecommendedImpactISK, got %v", est.RecommendedImpactISK)
	}
	// 5000 > 1% of 10000 → should use sqrt model
	if est.RecommendedImpactPct != est.SqrtImpactPct {
		t.Errorf("for large Q, recommended should equal sqrt, got recommended=%v sqrt=%v",
			est.RecommendedImpactPct, est.SqrtImpactPct)
	}
	if est.OptimalSlices < 1 {
		t.Errorf("OptimalSlices want >= 1, got %d", est.OptimalSlices)
	}
}

func TestEstimateImpact_SmallOrder(t *testing.T) {
	params := ImpactParams{
		Amihud:         0.00001,
		Sigma:          0.02,
		AvgDailyVolume: 100000, // 100k daily volume
		Valid:          true,
	}
	// Q=50 is 0.05% of daily volume → should use linear model
	est := EstimateImpact(params, 50, 100.0)
	if est.RecommendedImpactPct != est.LinearImpactPct {
		t.Errorf("for small Q, recommended should equal linear, got recommended=%v linear=%v",
			est.RecommendedImpactPct, est.LinearImpactPct)
	}
}
