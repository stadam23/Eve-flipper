package engine

import "math"

func buildStationForecast(row *StationCommandRow) StationCommandForecast {
	if row == nil {
		return StationCommandForecast{}
	}

	baseVolume := math.Max(0, float64(row.Trade.DailyVolume))
	if baseVolume <= 0 {
		baseVolume = math.Max(row.Trade.S2BPerDay, row.Trade.BfSPerDay)
	}
	if baseVolume <= 0 {
		baseVolume = math.Max(float64(row.Trade.BuyOrderCount+row.Trade.SellOrderCount), 1)
	}

	baseProfit := row.Trade.DailyProfit
	if baseProfit == 0 {
		baseProfit = row.Trade.RealizableDailyProfit
	}
	if baseProfit == 0 {
		baseProfit = row.Trade.TheoreticalDailyProfit
	}

	targetUnits := math.Max(1, math.Min(float64(row.Trade.BuyVolume), float64(row.Trade.SellVolume))*0.02)
	if row.OpenPositionQty > 0 {
		targetUnits = math.Max(targetUnits, float64(row.OpenPositionQty))
	}
	if row.ActiveOrderAtStation > 0 {
		targetUnits = math.Max(targetUnits, float64(row.Trade.BuyOrderCount+row.Trade.SellOrderCount))
	}
	etaBase := targetUnits / math.Max(baseVolume, 1)
	switch row.RecommendedAction {
	case StationActionCancel:
		etaBase = 0.15
	case StationActionReprice:
		etaBase = math.Max(0.25, etaBase)
	case StationActionNewEntry:
		etaBase = math.Max(0.50, etaBase)
	default:
		etaBase = math.Max(0.75, etaBase)
	}
	etaBase = clampRange(etaBase, 0.1, 30)

	uncertainty := stationForecastUncertainty(row)
	return StationCommandForecast{
		DailyVolume: stationForecastPositiveBand(baseVolume, uncertainty),
		DailyProfit: stationForecastProfitBand(baseProfit, uncertainty),
		ETADays:     stationForecastETABand(etaBase, uncertainty),
	}
}

func stationForecastUncertainty(row *StationCommandRow) float64 {
	confidence := row.Trade.ConfidenceScore
	if confidence <= 0 {
		switch row.Trade.ConfidenceLabel {
		case "high":
			confidence = 78
		case "medium":
			confidence = 56
		default:
			confidence = 34
		}
	}

	confPenalty := 1 - normalize(confidence, 0, 100)
	risk := 0.35*normalize(float64(row.Trade.CI), 0, 50) +
		0.35*normalize(float64(row.Trade.SDS), 0, 100) +
		0.15*normalize(row.Trade.DOS, 0, 60) +
		0.15*normalize(row.Trade.PVI, 0, 50)

	uncertainty := 0.18 + 0.52*confPenalty + 0.30*risk
	if !row.Trade.HistoryAvailable {
		uncertainty += 0.08
	}
	if row.RecommendedAction == StationActionCancel {
		uncertainty += 0.04
	}
	return clampRange(uncertainty, 0.12, 0.92)
}

func stationForecastPositiveBand(base, uncertainty float64) StationForecastBand {
	base = stationForecastSanitize(base)
	uncertainty = clampRange(uncertainty, 0.01, 1)
	if base <= 0 {
		return StationForecastBand{}
	}

	spread80 := clampRange(uncertainty*0.45, 0.03, 0.85)
	spread95 := clampRange(uncertainty*0.80, 0.06, 0.95)
	return StationForecastBand{
		P50: base,
		P80: stationForecastSanitize(base * (1 - spread80)),
		P95: stationForecastSanitize(base * (1 - spread95)),
	}
}

func stationForecastProfitBand(base, uncertainty float64) StationForecastBand {
	base = stationForecastSanitize(base)
	uncertainty = clampRange(uncertainty, 0.01, 1)
	if base == 0 {
		return StationForecastBand{}
	}

	spread80 := clampRange(uncertainty*0.45, 0.03, 0.85)
	spread95 := clampRange(uncertainty*0.80, 0.06, 0.95)
	if base > 0 {
		return StationForecastBand{
			P50: base,
			P80: stationForecastSanitize(base * (1 - spread80)),
			P95: stationForecastSanitize(base * (1 - spread95)),
		}
	}

	return StationForecastBand{
		P50: base,
		P80: stationForecastSanitize(base * (1 + spread80)),
		P95: stationForecastSanitize(base * (1 + spread95)),
	}
}

func stationForecastETABand(baseDays, uncertainty float64) StationForecastBand {
	baseDays = stationForecastSanitize(baseDays)
	if baseDays <= 0 {
		return StationForecastBand{}
	}

	spread80 := clampRange(uncertainty*0.55, 0.05, 0.95)
	spread95 := clampRange(uncertainty*0.95, 0.1, 1.2)
	return StationForecastBand{
		P50: baseDays,
		P80: stationForecastSanitize(baseDays * (1 + spread80)),
		P95: stationForecastSanitize(baseDays * (1 + spread95)),
	}
}

func stationForecastSanitize(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
