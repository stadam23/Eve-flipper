package engine

import (
	"sort"
	"time"

	"eve-flipper/internal/esi"
)

type StationCommandAction string

const (
	StationActionNewEntry StationCommandAction = "new_entry"
	StationActionReprice  StationCommandAction = "reprice"
	StationActionHold     StationCommandAction = "hold"
	StationActionCancel   StationCommandAction = "cancel"
)

type StationForecastBand struct {
	P50 float64 `json:"p50"`
	P80 float64 `json:"p80"`
	P95 float64 `json:"p95"`
}

type StationCommandForecast struct {
	DailyVolume StationForecastBand `json:"daily_volume"`
	DailyProfit StationForecastBand `json:"daily_profit"`
	ETADays     StationForecastBand `json:"eta_days"`
}

// StationCommandRow is one actionable row for the Station Command Center.
type StationCommandRow struct {
	Trade                    StationTrade           `json:"trade"`
	PersonalizedScore        float64                `json:"personalized_score"`
	RecommendedAction        StationCommandAction   `json:"recommended_action"`
	ActionReason             string                 `json:"action_reason"`
	Priority                 int                    `json:"priority"`
	ActiveOrderCount         int                    `json:"active_order_count"`
	ActiveOrderAtStation     int                    `json:"active_order_at_station"`
	OpenPositionQty          int64                  `json:"open_position_qty"`
	ExpectedDeltaDailyProfit float64                `json:"expected_delta_daily_profit"`
	Forecast                 StationCommandForecast `json:"forecast"`
}

// StationCommandSummary aggregates recommendation counts and context.
type StationCommandSummary struct {
	Rows              int `json:"rows"`
	NewEntryCount     int `json:"new_entry_count"`
	RepriceCount      int `json:"reprice_count"`
	HoldCount         int `json:"hold_count"`
	CancelCount       int `json:"cancel_count"`
	WithActiveOrders  int `json:"with_active_orders"`
	WithOpenPositions int `json:"with_open_positions"`
}

// StationCommandResult is the top-level recommendation payload.
type StationCommandResult struct {
	GeneratedAt string                `json:"generated_at"`
	Summary     StationCommandSummary `json:"summary"`
	Rows        []StationCommandRow   `json:"rows"`
}

type commandStationTypeKey struct {
	typeID     int32
	locationID int64
}

// BuildStationCommand converts raw station scan rows into an operator-oriented
// recommendation list using active orders and open inventory context.
func BuildStationCommand(trades []StationTrade, activeOrders []esi.CharacterOrder, openPositions []OpenPosition) StationCommandResult {
	activeByType := make(map[int32]int)
	activeByTypeStation := make(map[commandStationTypeKey]int)
	for _, o := range activeOrders {
		activeByType[o.TypeID]++
		activeByTypeStation[commandStationTypeKey{typeID: o.TypeID, locationID: o.LocationID}]++
	}

	openQtyByType := make(map[int32]int64)
	for _, pos := range openPositions {
		openQtyByType[pos.TypeID] += pos.Quantity
	}

	out := StationCommandResult{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Rows:        make([]StationCommandRow, 0, len(trades)),
	}

	for _, t := range trades {
		row := StationCommandRow{
			Trade:                    t,
			PersonalizedScore:        defaultStationCommandScore(t),
			RecommendedAction:        StationActionHold,
			ActionReason:             "pending action evaluation",
			Priority:                 stationCommandActionPriority(StationActionHold),
			ActiveOrderCount:         activeByType[t.TypeID],
			ActiveOrderAtStation:     activeByTypeStation[commandStationTypeKey{typeID: t.TypeID, locationID: t.StationID}],
			OpenPositionQty:          openQtyByType[t.TypeID],
			ExpectedDeltaDailyProfit: t.DailyProfit,
		}

		evaluateStationAction(&row)
		row.Forecast = buildStationForecast(&row)
		row.PersonalizedScore = clampRange(row.PersonalizedScore, 0, 100)

		if row.ActiveOrderCount > 0 {
			out.Summary.WithActiveOrders++
		}
		if row.OpenPositionQty > 0 {
			out.Summary.WithOpenPositions++
		}
		switch row.RecommendedAction {
		case StationActionNewEntry:
			out.Summary.NewEntryCount++
		case StationActionReprice:
			out.Summary.RepriceCount++
		case StationActionHold:
			out.Summary.HoldCount++
		case StationActionCancel:
			out.Summary.CancelCount++
		}
		out.Rows = append(out.Rows, row)
	}

	sort.Slice(out.Rows, func(i, j int) bool {
		if out.Rows[i].Priority != out.Rows[j].Priority {
			return out.Rows[i].Priority > out.Rows[j].Priority
		}
		if out.Rows[i].PersonalizedScore != out.Rows[j].PersonalizedScore {
			return out.Rows[i].PersonalizedScore > out.Rows[j].PersonalizedScore
		}
		if out.Rows[i].Trade.DailyProfit != out.Rows[j].Trade.DailyProfit {
			return out.Rows[i].Trade.DailyProfit > out.Rows[j].Trade.DailyProfit
		}
		if out.Rows[i].Trade.CTS != out.Rows[j].Trade.CTS {
			return out.Rows[i].Trade.CTS > out.Rows[j].Trade.CTS
		}
		if out.Rows[i].Trade.TypeID != out.Rows[j].Trade.TypeID {
			return out.Rows[i].Trade.TypeID < out.Rows[j].Trade.TypeID
		}
		return out.Rows[i].Trade.StationID < out.Rows[j].Trade.StationID
	})

	out.Summary.Rows = len(out.Rows)
	return out
}

func defaultStationCommandScore(t StationTrade) float64 {
	if t.CTS > 0 {
		return t.CTS
	}
	if t.ConfidenceScore > 0 {
		return t.ConfidenceScore
	}
	return t.MarginPercent
}

func clampRange(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
