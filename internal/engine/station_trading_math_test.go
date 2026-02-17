package engine

import "testing"

func TestEstimateSellUnitsPerDay_AllowsBvSBelowOne(t *testing.T) {
	daily := 100.0
	buyVol := int32(500)
	sellVol := int32(1000)

	sellPerDay := estimateSellUnitsPerDay(daily, buyVol, sellVol)
	if sellPerDay <= daily {
		t.Fatalf("sellPerDay = %v, want > %v", sellPerDay, daily)
	}

	bvs := daily / sellPerDay
	if bvs >= 1 {
		t.Fatalf("BvS = %v, want < 1", bvs)
	}
}

func TestEstimateSellUnitsPerDay_AllowsBvSAboveOne(t *testing.T) {
	daily := 100.0
	buyVol := int32(1000)
	sellVol := int32(500)

	sellPerDay := estimateSellUnitsPerDay(daily, buyVol, sellVol)
	if sellPerDay >= daily {
		t.Fatalf("sellPerDay = %v, want < %v", sellPerDay, daily)
	}

	bvs := daily / sellPerDay
	if bvs <= 1 {
		t.Fatalf("BvS = %v, want > 1", bvs)
	}
}

func TestStationExecutionDesiredQty(t *testing.T) {
	if got := stationExecutionDesiredQty(400, 1000, 300); got != 300 {
		t.Fatalf("desired qty with dailyShare cap = %d, want 300", got)
	}
	if got := stationExecutionDesiredQty(50, 40, 100); got != 40 {
		t.Fatalf("desired qty with depth cap = %d, want 40", got)
	}
	if got := stationExecutionDesiredQty(0, 5000, 8000); got != 1000 {
		t.Fatalf("fallback desired qty = %d, want 1000", got)
	}
	if got := stationExecutionDesiredQty(0, 0, 10); got != 0 {
		t.Fatalf("zero depth desired qty = %d, want 0", got)
	}
}

func TestStationExecutionDesiredQtyFromDailyShare(t *testing.T) {
	if got := stationExecutionDesiredQtyFromDailyShare(0, 5000, 8000); got != 0 {
		t.Fatalf("strict daily-share qty with unknown share = %d, want 0", got)
	}
	if got := stationExecutionDesiredQtyFromDailyShare(120, 1000, 90); got != 90 {
		t.Fatalf("strict daily-share qty with depth cap = %d, want 90", got)
	}
}

func TestEstimateSideFlowsPerDay_MassBalance(t *testing.T) {
	total := 100.0
	s2b, bfs := estimateSideFlowsPerDay(total, 600, 400)
	if s2b <= 0 || bfs <= 0 {
		t.Fatalf("flows should be positive, got s2b=%v bfs=%v", s2b, bfs)
	}
	gotTotal := s2b + bfs
	if gotTotal != total {
		t.Fatalf("mass-balance violated: s2b+bfs=%v, want %v", gotTotal, total)
	}

	halfS2B, halfBfS := estimateSideFlowsPerDay(total, 0, 0)
	if halfS2B != 50 || halfBfS != 50 {
		t.Fatalf("zero-depth split = %v/%v, want 50/50", halfS2B, halfBfS)
	}
}

func TestApplyStationTradeFilters_UsesExecutionAwareMarginsAndHistory(t *testing.T) {
	rows := []StationTrade{
		{
			TypeID:            1,
			MarginPercent:     20,
			RealMarginPercent: 2,
			ProfitPerUnit:     10_000,
			RealProfit:        1000,
			FilledQty:         1000,
			DailyVolume:       200,
			S2BPerDay:         100,
			BfSPerDay:         100,
			S2BBfSRatio:       1,
			HistoryAvailable:  true,
		},
	}
	params := StationTradeParams{
		MinMargin:      5,
		MinDailyVolume: 10,
	}
	out := applyStationTradeFilters(rows, params)
	if len(out) != 0 {
		t.Fatalf("expected row to be dropped by RealMarginPercent, got %d rows", len(out))
	}

	rows[0].RealMarginPercent = 8
	rows[0].HistoryAvailable = false
	out = applyStationTradeFilters(rows, params)
	if len(out) != 0 {
		t.Fatalf("expected row to be dropped by missing history, got %d rows", len(out))
	}

	rows[0].HistoryAvailable = true
	rows[0].RealMarginPercent = -1
	rows[0].FilledQty = 100
	rows[0].RealProfit = -500
	params.MinItemProfit = 1
	out = applyStationTradeFilters(rows, params)
	if len(out) != 0 {
		t.Fatalf("expected row to be dropped by execution-aware negative margin/profit, got %d rows", len(out))
	}

	rows[0].RealMarginPercent = 8
	rows[0].RealProfit = 100
	rows[0].FilledQty = 0
	rows[0].DailyVolume = 50
	out = applyStationTradeFilters(rows, StationTradeParams{})
	if len(out) != 0 {
		t.Fatalf("expected row to be dropped as unexecutable with positive history flow, got %d rows", len(out))
	}
}
