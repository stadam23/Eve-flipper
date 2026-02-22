package engine

import (
	"testing"

	"eve-flipper/internal/esi"
)

func TestBuildStationCommand_ActionSelection(t *testing.T) {
	trades := []StationTrade{
		{
			TypeID:          34,
			StationID:       60003760,
			CTS:             72,
			DailyProfit:     1_000_000,
			MarginPercent:   12,
			ConfidenceLabel: "high",
		},
		{
			TypeID:          35,
			StationID:       60003760,
			CTS:             65,
			DailyProfit:     700_000,
			MarginPercent:   9,
			ConfidenceLabel: "low",
			CI:              40,
		},
		{
			TypeID:            36,
			StationID:         60008494,
			CTS:               55,
			DailyProfit:       -100_000,
			MarginPercent:     3,
			RealMarginPercent: -2,
		},
	}
	activeOrders := []esi.CharacterOrder{
		{TypeID: 35, LocationID: 60003760},
		{TypeID: 36, LocationID: 60008494},
	}
	openPositions := []OpenPosition{
		{TypeID: 34, Quantity: 120},
	}

	got := BuildStationCommand(trades, activeOrders, openPositions)

	if got.Summary.Rows != 3 {
		t.Fatalf("summary rows = %d, want 3", got.Summary.Rows)
	}
	if got.Summary.RepriceCount == 0 {
		t.Fatalf("expected at least one reprice recommendation")
	}
	if got.Summary.CancelCount != 1 {
		t.Fatalf("cancel count = %d, want 1", got.Summary.CancelCount)
	}

	byType := make(map[int32]StationCommandRow)
	for _, row := range got.Rows {
		byType[row.Trade.TypeID] = row
	}

	if row := byType[34]; row.RecommendedAction != StationActionReprice {
		t.Fatalf("type 34 action = %q, want reprice (inventory-aware)", row.RecommendedAction)
	}
	if row := byType[35]; row.RecommendedAction != StationActionReprice {
		t.Fatalf("type 35 action = %q, want reprice", row.RecommendedAction)
	}
	if row := byType[35]; row.ActionReason == "" || row.Priority <= 0 {
		t.Fatalf("type 35 should have non-empty reason and priority, got reason=%q priority=%d", row.ActionReason, row.Priority)
	}
	if row := byType[36]; row.RecommendedAction != StationActionCancel {
		t.Fatalf("type 36 action = %q, want cancel", row.RecommendedAction)
	}
	if row := byType[34]; row.Forecast.DailyVolume.P50 <= 0 {
		t.Fatalf("type 34 forecast daily volume p50 = %v, want > 0", row.Forecast.DailyVolume.P50)
	}
	if row := byType[34]; !(row.Forecast.DailyProfit.P50 >= row.Forecast.DailyProfit.P80 && row.Forecast.DailyProfit.P80 >= row.Forecast.DailyProfit.P95) {
		t.Fatalf("type 34 positive profit forecast should be p50>=p80>=p95, got %v/%v/%v",
			row.Forecast.DailyProfit.P50, row.Forecast.DailyProfit.P80, row.Forecast.DailyProfit.P95)
	}
	if row := byType[36]; !(row.Forecast.DailyProfit.P50 >= row.Forecast.DailyProfit.P80 && row.Forecast.DailyProfit.P80 >= row.Forecast.DailyProfit.P95) {
		t.Fatalf("type 36 negative profit forecast should be p50>=p80>=p95 (more conservative), got %v/%v/%v",
			row.Forecast.DailyProfit.P50, row.Forecast.DailyProfit.P80, row.Forecast.DailyProfit.P95)
	}
	if row := byType[35]; !(row.Forecast.ETADays.P50 <= row.Forecast.ETADays.P80 && row.Forecast.ETADays.P80 <= row.Forecast.ETADays.P95) {
		t.Fatalf("type 35 eta forecast should be p50<=p80<=p95, got %v/%v/%v",
			row.Forecast.ETADays.P50, row.Forecast.ETADays.P80, row.Forecast.ETADays.P95)
	}
}

func TestBuildStationCommand_SortingByPriorityThenScore(t *testing.T) {
	trades := []StationTrade{
		{
			TypeID:            1001,
			StationID:         60000001,
			CTS:               80,
			DailyProfit:       -10,
			RealMarginPercent: -1,
		},
		{
			TypeID:        1002,
			StationID:     60000002,
			CTS:           90,
			DailyProfit:   200,
			MarginPercent: 5,
		},
	}
	activeOrders := []esi.CharacterOrder{
		{TypeID: 1001, LocationID: 60000001},
	}

	got := BuildStationCommand(trades, activeOrders, nil)
	if len(got.Rows) != 2 {
		t.Fatalf("rows len = %d, want 2", len(got.Rows))
	}
	if got.Rows[0].RecommendedAction != StationActionCancel {
		t.Fatalf("first action = %q, want cancel (highest priority)", got.Rows[0].RecommendedAction)
	}
}
