package engine

import "testing"

func TestBuildStationForecast_BandOrdering(t *testing.T) {
	row := &StationCommandRow{
		Trade: StationTrade{
			DailyVolume:            120,
			DailyProfit:            1_200_000,
			ConfidenceScore:        82,
			ConfidenceLabel:        "high",
			HistoryAvailable:       true,
			DOS:                    6,
			PVI:                    8,
			SDS:                    12,
			CI:                     8,
			BuyOrderCount:          12,
			SellOrderCount:         10,
			BuyVolume:              300,
			SellVolume:             280,
			TheoreticalDailyProfit: 1_250_000,
		},
		RecommendedAction: StationActionReprice,
	}

	f := buildStationForecast(row)
	if !(f.DailyVolume.P50 >= f.DailyVolume.P80 && f.DailyVolume.P80 >= f.DailyVolume.P95) {
		t.Fatalf("daily volume band order invalid: %+v", f.DailyVolume)
	}
	if !(f.DailyProfit.P50 >= f.DailyProfit.P80 && f.DailyProfit.P80 >= f.DailyProfit.P95) {
		t.Fatalf("daily profit band order invalid: %+v", f.DailyProfit)
	}
	if !(f.ETADays.P50 <= f.ETADays.P80 && f.ETADays.P80 <= f.ETADays.P95) {
		t.Fatalf("eta band order invalid: %+v", f.ETADays)
	}
}

func TestBuildStationForecast_UncertaintyWiderForLowConfidence(t *testing.T) {
	high := &StationCommandRow{
		Trade: StationTrade{
			DailyVolume:      100,
			DailyProfit:      900_000,
			ConfidenceScore:  85,
			ConfidenceLabel:  "high",
			HistoryAvailable: true,
			DOS:              8,
			PVI:              10,
			SDS:              15,
			CI:               12,
			BuyVolume:        220,
			SellVolume:       200,
		},
		RecommendedAction: StationActionReprice,
	}
	low := &StationCommandRow{
		Trade: StationTrade{
			DailyVolume:      100,
			DailyProfit:      900_000,
			ConfidenceScore:  22,
			ConfidenceLabel:  "low",
			HistoryAvailable: false,
			DOS:              36,
			PVI:              28,
			SDS:              62,
			CI:               39,
			BuyVolume:        220,
			SellVolume:       200,
		},
		RecommendedAction: StationActionReprice,
	}

	fHigh := buildStationForecast(high)
	fLow := buildStationForecast(low)

	highSpread := fHigh.DailyProfit.P50 - fHigh.DailyProfit.P95
	lowSpread := fLow.DailyProfit.P50 - fLow.DailyProfit.P95
	if lowSpread <= highSpread {
		t.Fatalf("expected wider low-confidence band, high=%v low=%v", highSpread, lowSpread)
	}

	highETAStretch := fHigh.ETADays.P95 - fHigh.ETADays.P50
	lowETAStretch := fLow.ETADays.P95 - fLow.ETADays.P50
	if lowETAStretch <= highETAStretch {
		t.Fatalf("expected wider low-confidence eta stretch, high=%v low=%v", highETAStretch, lowETAStretch)
	}
}
