package engine

import (
	"context"
	"errors"
	"math"
	"testing"

	"eve-flipper/internal/esi"
	"eve-flipper/internal/sde"
)

// getContractFilters returns effective minPrice, maxMargin, minPricedRatio (defaults when params are 0).

func TestGetContractFilters_Defaults(t *testing.T) {
	var params ScanParams
	minPrice, maxMargin, minPricedRatio := getContractFilters(params)
	if minPrice != DefaultMinContractPrice {
		t.Errorf("minPrice = %v, want DefaultMinContractPrice %v", minPrice, DefaultMinContractPrice)
	}
	if maxMargin != DefaultMaxContractMargin {
		t.Errorf("maxMargin = %v, want DefaultMaxContractMargin %v", maxMargin, DefaultMaxContractMargin)
	}
	if minPricedRatio != DefaultMinPricedRatio {
		t.Errorf("minPricedRatio = %v, want DefaultMinPricedRatio %v", minPricedRatio, DefaultMinPricedRatio)
	}
}

func TestGetContractFilters_Explicit(t *testing.T) {
	params := ScanParams{
		MinContractPrice:  50_000_000,
		MaxContractMargin: 80,
		MinPricedRatio:    0.9,
	}
	minPrice, maxMargin, minPricedRatio := getContractFilters(params)
	if minPrice != 50_000_000 {
		t.Errorf("minPrice = %v, want 50_000_000", minPrice)
	}
	if maxMargin != 80 {
		t.Errorf("maxMargin = %v, want 80", maxMargin)
	}
	if minPricedRatio != 0.9 {
		t.Errorf("minPricedRatio = %v, want 0.9", minPricedRatio)
	}
}

func TestGetContractFilters_PartialDefaults(t *testing.T) {
	// Only MinContractPrice set; others use defaults
	params := ScanParams{MinContractPrice: 1_000_000}
	minPrice, maxMargin, minPricedRatio := getContractFilters(params)
	if minPrice != 1_000_000 {
		t.Errorf("minPrice = %v, want 1_000_000", minPrice)
	}
	if maxMargin != DefaultMaxContractMargin {
		t.Errorf("maxMargin = %v, want default %v", maxMargin, DefaultMaxContractMargin)
	}
	if minPricedRatio != DefaultMinPricedRatio {
		t.Errorf("minPricedRatio = %v, want default %v", minPricedRatio, DefaultMinPricedRatio)
	}
}

func TestGetContractFilters_ZeroMaxMarginUsesDefault(t *testing.T) {
	params := ScanParams{MaxContractMargin: 0}
	_, maxMargin, _ := getContractFilters(params)
	if maxMargin != DefaultMaxContractMargin {
		t.Errorf("maxMargin when 0 = %v, want default %v", maxMargin, DefaultMaxContractMargin)
	}
}

func TestGetContractFilters_MinPricedRatioPercentAndClamp(t *testing.T) {
	params := ScanParams{MinPricedRatio: 80} // accidental percent input from API client
	_, _, minPricedRatio := getContractFilters(params)
	if minPricedRatio != 0.8 {
		t.Errorf("minPricedRatio(80) = %v, want 0.8", minPricedRatio)
	}

	params = ScanParams{MinPricedRatio: 0.01}
	_, _, minPricedRatio = getContractFilters(params)
	if minPricedRatio != 0.1 {
		t.Errorf("minPricedRatio lower clamp = %v, want 0.1", minPricedRatio)
	}
}

func TestContractSellValueMultiplier_InstantLiquidation_IgnoresBroker(t *testing.T) {
	params := ScanParams{
		SalesTaxPercent:            8,
		BrokerFeePercent:           3,
		ContractInstantLiquidation: true,
	}
	got := contractSellValueMultiplier(params)
	want := 0.92 // 1 - 8%
	if got != want {
		t.Errorf("contractSellValueMultiplier instant = %v, want %v", got, want)
	}
}

func TestContractSellValueMultiplier_Estimate_IncludesBroker(t *testing.T) {
	params := ScanParams{
		SalesTaxPercent:            8,
		BrokerFeePercent:           3,
		ContractInstantLiquidation: false,
	}
	got := contractSellValueMultiplier(params)
	want := 0.89 // 1 - (8% + 3%)
	if got != want {
		t.Errorf("contractSellValueMultiplier estimate = %v, want %v", got, want)
	}
}

func TestContractSellValueMultiplier_SplitUsesSellSideOnly(t *testing.T) {
	params := ScanParams{
		SplitTradeFees:             true,
		BuyBrokerFeePercent:        0.5,
		SellBrokerFeePercent:       0.2,
		BuySalesTaxPercent:         0,
		SellSalesTaxPercent:        3.6,
		ContractInstantLiquidation: false,
	}
	got := contractSellValueMultiplier(params)
	want := 0.962 // 1 - (3.6% + 0.2%)
	if got != want {
		t.Errorf("contractSellValueMultiplier split estimate = %v, want %v", got, want)
	}

	params.ContractInstantLiquidation = true
	got = contractSellValueMultiplier(params)
	want = 0.964 // 1 - 3.6%
	if got != want {
		t.Errorf("contractSellValueMultiplier split instant = %v, want %v", got, want)
	}
}

func TestContractHoldDays_DefaultAndClamp(t *testing.T) {
	if got := contractHoldDays(ScanParams{}); got != DefaultContractHoldDays {
		t.Errorf("contractHoldDays default = %d, want %d", got, DefaultContractHoldDays)
	}
	if got := contractHoldDays(ScanParams{ContractHoldDays: 365}); got != 180 {
		t.Errorf("contractHoldDays clamp = %d, want 180", got)
	}
}

func TestContractTargetConfidence_DefaultAndClamp(t *testing.T) {
	if got := contractTargetConfidence(ScanParams{}); got != DefaultContractTargetConfidence {
		t.Errorf("contractTargetConfidence default = %v, want %v", got, DefaultContractTargetConfidence)
	}
	if got := contractTargetConfidence(ScanParams{ContractTargetConfidence: 140}); got != 100 {
		t.Errorf("contractTargetConfidence clamp = %v, want 100", got)
	}
}

func TestEstimateFillDaysAndProbability(t *testing.T) {
	fillDays := estimateFillDays(350, 100) // effective/day = 35
	if fillDays != 10 {
		t.Errorf("estimateFillDays = %v, want 10", fillDays)
	}
	p := fillProbabilityWithinDays(fillDays, 7)
	if p <= 0 || p >= 1 {
		t.Errorf("fillProbabilityWithinDays = %v, want (0,1)", p)
	}
	// No volume => impossible fill in model.
	if p0 := fillProbabilityWithinDays(estimateFillDays(10, 0), 7); p0 != 0 {
		t.Errorf("fillProbabilityWithinDays(no volume) = %v, want 0", p0)
	}
}

func TestEstimateFillDays_Monotone(t *testing.T) {
	fastMarket := estimateFillDays(100, 1_000)
	slowMarket := estimateFillDays(100, 100)
	if !(fastMarket < slowMarket) {
		t.Fatalf("fill days should decrease with higher daily volume: fast=%f slow=%f", fastMarket, slowMarket)
	}

	smallOrder := estimateFillDays(100, 500)
	largeOrder := estimateFillDays(500, 500)
	if !(smallOrder < largeOrder) {
		t.Fatalf("fill days should increase with quantity: small=%f large=%f", smallOrder, largeOrder)
	}
}

func TestFillProbabilityWithinDays_MonotoneAndBounded(t *testing.T) {
	for _, fillDays := range []float64{1, 3, 7, 30, math.Inf(1)} {
		prev := -1.0
		for _, horizon := range []float64{1, 3, 7, 14, 30} {
			p := fillProbabilityWithinDays(fillDays, horizon)
			if p < 0 || p > 1 {
				t.Fatalf("probability out of bounds: fillDays=%f horizon=%f p=%f", fillDays, horizon, p)
			}
			if prev > p {
				t.Fatalf("probability should be non-decreasing with horizon: prev=%f cur=%f", prev, p)
			}
			prev = p
		}
	}

	short := fillProbabilityWithinDays(5, 7)
	long := fillProbabilityWithinDays(20, 7)
	if !(short > long) {
		t.Fatalf("probability should decrease with slower fill: short=%f long=%f", short, long)
	}
}

func TestContractCarryDays(t *testing.T) {
	if got := contractCarryDays(7, 0); got != 7 {
		t.Fatalf("contractCarryDays(7,0) = %f, want 7", got)
	}
	if got := contractCarryDays(7, 3.5); got != 3.5 {
		t.Fatalf("contractCarryDays(7,3.5) = %f, want 3.5", got)
	}
	if got := contractCarryDays(7, 20); got != 7 {
		t.Fatalf("contractCarryDays(7,20) = %f, want 7", got)
	}
	if got := contractCarryDays(0, 3); got != 0 {
		t.Fatalf("contractCarryDays(0,3) = %f, want 0", got)
	}
}

func TestIsMarketDisabledType(t *testing.T) {
	if !isMarketDisabledType(MPTCTypeID) {
		t.Fatalf("MPTCTypeID should be marked market-disabled")
	}
	if isMarketDisabledType(PLEXTypeID) {
		t.Fatalf("PLEXTypeID should not be market-disabled")
	}
}

func TestSelectInstantLiquidationSystem_RequiresSingleSystemForAllTypes(t *testing.T) {
	items := []instantValuationItem{
		{TypeID: 1, Quantity: 1, Label: "Item A", ValueFactor: 1},
		{TypeID: 2, Quantity: 1, Label: "Item B", ValueFactor: 1},
	}
	books := map[int32]map[int32][]esi.MarketOrder{
		1: {
			30000142: {{TypeID: 1, SystemID: 30000142, IsBuyOrder: true, Price: 100, VolumeRemain: 10}},
		},
		2: {
			30002187: {{TypeID: 2, SystemID: 30002187, IsBuyOrder: true, Price: 200, VolumeRemain: 10}},
		},
	}

	if _, ok := selectInstantLiquidationSystem(items, books, nil); ok {
		t.Fatalf("expected no valid system when item liquidity is split across different systems")
	}
}

func TestSelectInstantLiquidationSystem_PicksBestCommonSystem(t *testing.T) {
	items := []instantValuationItem{
		{TypeID: 1, Quantity: 1, Label: "Item A", ValueFactor: 1},
		{TypeID: 2, Quantity: 1, Label: "Item B", ValueFactor: 1},
	}
	books := map[int32]map[int32][]esi.MarketOrder{
		1: {
			30000142: {{TypeID: 1, SystemID: 30000142, IsBuyOrder: true, Price: 100, VolumeRemain: 10}},
			30002187: {{TypeID: 1, SystemID: 30002187, IsBuyOrder: true, Price: 95, VolumeRemain: 10}},
		},
		2: {
			30000142: {{TypeID: 2, SystemID: 30000142, IsBuyOrder: true, Price: 40, VolumeRemain: 10}},
			30002187: {{TypeID: 2, SystemID: 30002187, IsBuyOrder: true, Price: 80, VolumeRemain: 10}},
		},
	}

	choice, ok := selectInstantLiquidationSystem(items, books, nil)
	if !ok {
		t.Fatalf("expected a valid liquidation system")
	}
	if choice.SystemID != 30002187 {
		t.Fatalf("SystemID = %d, want 30002187", choice.SystemID)
	}
	if math.Abs(choice.MarketValue-175) > 1e-9 {
		t.Fatalf("MarketValue = %f, want 175", choice.MarketValue)
	}
	if choice.PricedCount != 2 {
		t.Fatalf("PricedCount = %d, want 2", choice.PricedCount)
	}
}

func TestSelectInstantLiquidationSystem_RespectsAllowedSystems(t *testing.T) {
	items := []instantValuationItem{
		{TypeID: 1, Quantity: 1, Label: "Item A", ValueFactor: 1},
	}
	books := map[int32]map[int32][]esi.MarketOrder{
		1: {
			30000142: {{TypeID: 1, SystemID: 30000142, IsBuyOrder: true, Price: 200, VolumeRemain: 10}},
			30002187: {{TypeID: 1, SystemID: 30002187, IsBuyOrder: true, Price: 150, VolumeRemain: 10}},
		},
	}

	allowOnlyLowsec := func(systemID int32) bool { return systemID != 30000142 }
	choice, ok := selectInstantLiquidationSystem(items, books, allowOnlyLowsec)
	if !ok {
		t.Fatalf("expected a valid liquidation system after filtering")
	}
	if choice.SystemID != 30002187 {
		t.Fatalf("SystemID = %d, want 30002187", choice.SystemID)
	}
}

func TestIsHighsecRestrictedShipGroup(t *testing.T) {
	if !isHighsecRestrictedShipGroup(883, "Capital Industrial Ship") {
		t.Fatalf("group 883 (Capital Industrial Ship) must be highsec-restricted")
	}
	if !isHighsecRestrictedShipGroup(0, "Carrier") {
		t.Fatalf("carrier name fallback must be highsec-restricted")
	}
	if isHighsecRestrictedShipGroup(902, "Jump Freighter") {
		t.Fatalf("jump freighter should not be highsec-restricted")
	}
}

func TestShouldExcludeRigWithShip(t *testing.T) {
	if !shouldExcludeRigWithShip(esi.ContractItem{Flag: 92}, "Large Core Defense Field Extender I", 3, false) {
		t.Fatalf("flagged rig slot should be excluded")
	}
	if !shouldExcludeRigWithShip(esi.ContractItem{Singleton: true}, "Large Core Defense Field Extender I", 3, false) {
		t.Fatalf("singleton rig should be excluded")
	}
	if !shouldExcludeRigWithShip(esi.ContractItem{}, "Large Core Defense Field Extender I", 3, false) {
		t.Fatalf("size-matched rig should be excluded")
	}
	if shouldExcludeRigWithShip(esi.ContractItem{}, "Small Core Defense Field Extender I", 3, false) {
		t.Fatalf("mismatched rig without fitted signal should not be excluded")
	}
	if !shouldExcludeRigWithShip(esi.ContractItem{}, "Small Core Defense Field Extender I", 3, true) {
		t.Fatalf("forceExclude should exclude all rigs when ship is present")
	}
	if shouldExcludeRigWithShip(esi.ContractItem{Flag: 92}, "Large Core Defense Field Extender I", 0, true) {
		t.Fatalf("no ship present: rig should not be excluded by ship logic")
	}
}

func TestBlockedContractTypeID(t *testing.T) {
	items := []esi.ContractItem{
		{TypeID: 34, Quantity: 100},        // Tritanium
		{TypeID: MPTCTypeID, Quantity: 1},  // market-disabled
		{TypeID: PLEXTypeID, Quantity: 1},  // tradable
		{TypeID: MPTCTypeID, Quantity: 0},  // ignored (non-positive qty)
		{TypeID: MPTCTypeID, Quantity: -5}, // ignored (non-positive qty)
	}
	if got := blockedContractTypeID(items); got != MPTCTypeID {
		t.Fatalf("blockedContractTypeID = %d, want %d", got, MPTCTypeID)
	}
	if got := blockedContractTypeID([]esi.ContractItem{{TypeID: 34, Quantity: 10}}); got != 0 {
		t.Fatalf("blockedContractTypeID(non-blocked) = %d, want 0", got)
	}
}

func TestScanContractsWithContext_CanceledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := &Scanner{}
	_, err := s.ScanContractsWithContext(ctx, ScanParams{}, func(string) {})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestContractItemLabel_UsesSDENameFirst(t *testing.T) {
	s := &Scanner{
		SDE: &sde.Data{
			Types: map[int32]*sde.ItemType{
				34: {ID: 34, Name: "Tritanium"},
			},
		},
	}
	got := s.contractItemLabel(34, map[int32]string{})
	if got != "Tritanium" {
		t.Fatalf("contractItemLabel = %q, want Tritanium", got)
	}
}

func TestContractItemLabel_FallbackToTypeIDAndCachesPerScan(t *testing.T) {
	s := &Scanner{SDE: &sde.Data{Types: map[int32]*sde.ItemType{}}}
	cache := map[int32]string{}

	got := s.contractItemLabel(34133, cache)
	if got != "Type 34133" {
		t.Fatalf("contractItemLabel = %q, want Type 34133", got)
	}
	if cache[34133] != "Type 34133" {
		t.Fatalf("expected scan cache to store fallback label, got %q", cache[34133])
	}
}
