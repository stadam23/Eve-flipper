package engine

import "math"

const stationActionModelVersion = "v2"

type stationActionDecision struct {
	Action           StationCommandAction
	Reason           string
	ExpectedDeltaPct float64
	Priority         int
	ScoreDelta       float64
}

func evaluateStationAction(row *StationCommandRow) {
	decision := stationActionDecision{
		Action:           StationActionHold,
		Reason:           "no prioritized action",
		ExpectedDeltaPct: 0,
		Priority:         stationCommandActionPriority(StationActionHold),
	}

	riskPenalty := stationRiskPenalty(row)
	scoreDelta := -riskPenalty

	if row.ActiveOrderAtStation > 0 {
		switch {
		case row.Trade.DailyProfit <= 0 || row.Trade.RealMarginPercent < 0:
			decision = stationActionDecision{
				Action:           StationActionCancel,
				Reason:           "active order at station has negative expected edge",
				ExpectedDeltaPct: -1.0,
				Priority:         stationCommandActionPriority(StationActionCancel),
				ScoreDelta:       scoreDelta - 20,
			}
		case row.Trade.ConfidenceLabel == "low" || row.Trade.CI >= 35 || row.Trade.DOS >= 45:
			decision = stationActionDecision{
				Action:           StationActionReprice,
				Reason:           "active order at station with weak confidence or queue pressure",
				ExpectedDeltaPct: 0.20,
				Priority:         stationCommandActionPriority(StationActionReprice),
				ScoreDelta:       scoreDelta - 6,
			}
		default:
			decision = stationActionDecision{
				Action:           StationActionHold,
				Reason:           "active order at station remains healthy",
				ExpectedDeltaPct: 0.05,
				Priority:         stationCommandActionPriority(StationActionHold),
				ScoreDelta:       scoreDelta + 4,
			}
		}
		applyStationDecision(row, decision)
		return
	}

	if row.ActiveOrderCount > 0 {
		if row.Trade.DailyProfit > 0 && row.Trade.ConfidenceLabel != "low" {
			decision = stationActionDecision{
				Action:           StationActionReprice,
				Reason:           "active order exists on other station; reprice/move context",
				ExpectedDeltaPct: 0.10,
				Priority:         stationCommandActionPriority(StationActionReprice),
				ScoreDelta:       scoreDelta - 2,
			}
		} else {
			decision = stationActionDecision{
				Action:           StationActionHold,
				Reason:           "active order exists but signal is weak",
				ExpectedDeltaPct: 0,
				Priority:         stationCommandActionPriority(StationActionHold),
				ScoreDelta:       scoreDelta - 8,
			}
		}
		applyStationDecision(row, decision)
		return
	}

	if row.Trade.DailyProfit <= 0 {
		decision = stationActionDecision{
			Action:           StationActionHold,
			Reason:           "no positive daily edge after filters",
			ExpectedDeltaPct: 0,
			Priority:         stationCommandActionPriority(StationActionHold),
			ScoreDelta:       scoreDelta - 16,
		}
		applyStationDecision(row, decision)
		return
	}

	if row.OpenPositionQty > 0 {
		decision = stationActionDecision{
			Action:           StationActionReprice,
			Reason:           "inventory available; prioritize relist before new buy orders",
			ExpectedDeltaPct: 0.35,
			Priority:         stationCommandActionPriority(StationActionReprice),
			ScoreDelta:       scoreDelta + 8,
		}
		applyStationDecision(row, decision)
		return
	}

	decision = stationActionDecision{
		Action:           StationActionNewEntry,
		Reason:           "no active orders and positive signal",
		ExpectedDeltaPct: 1.0,
		Priority:         stationCommandActionPriority(StationActionNewEntry),
		ScoreDelta:       scoreDelta + 10,
	}
	applyStationDecision(row, decision)
}

func stationRiskPenalty(row *StationCommandRow) float64 {
	penalty := 0.0
	if row.Trade.ConfidenceLabel == "low" {
		penalty += 10
	}
	if row.Trade.SDS >= 50 {
		penalty += 14
	} else if row.Trade.SDS >= 35 {
		penalty += 6
	}
	if row.Trade.CI >= 35 {
		penalty += 8
	} else if row.Trade.CI >= 20 {
		penalty += 4
	}
	if row.Trade.DOS >= 45 {
		penalty += 8
	} else if row.Trade.DOS >= 25 {
		penalty += 4
	}
	if row.Trade.PVI >= 35 {
		penalty += 6
	}
	return penalty
}

func applyStationDecision(row *StationCommandRow, decision stationActionDecision) {
	row.RecommendedAction = decision.Action
	row.ActionReason = decision.Reason + " (" + stationActionModelVersion + ")"
	row.Priority = decision.Priority
	row.PersonalizedScore = row.PersonalizedScore + decision.ScoreDelta
	row.ExpectedDeltaDailyProfit = sanitizeStationDelta(row.Trade.DailyProfit * decision.ExpectedDeltaPct)
}

func stationCommandActionPriority(a StationCommandAction) int {
	switch a {
	case StationActionCancel:
		return 100
	case StationActionReprice:
		return 85
	case StationActionNewEntry:
		return 70
	default:
		return 40
	}
}

func sanitizeStationDelta(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
