package engine

import (
	"math"
	"sort"
	"time"

	"eve-flipper/internal/esi"
)

// PortfolioRiskSummary is a user-friendly snapshot of portfolio risk for a character.
// All fields are in ISK or simple scores; no financial jargon is exposed to the UI.
type PortfolioRiskSummary struct {
	// RiskScore is 0-100 (higher = more risk).
	RiskScore float64 `json:"risk_score"`
	// RiskLevel is a coarse label: "safe", "balanced", "high".
	RiskLevel string `json:"risk_level"`

	// One-day worst loss levels (historical, 95/99%).
	Var95 float64 `json:"var_95"`
	Var99 float64 `json:"var_99"`
	// Expected shortfall (average of worst X% days).
	ES95 float64 `json:"es_95"`
	ES99 float64 `json:"es_99"`

	// TypicalDailyPnl is a robust scale of recent daily PnL.
	TypicalDailyPnl float64 `json:"typical_daily_pnl"`
	// WorstDayLoss is the worst historical daily loss in ISK.
	WorstDayLoss float64 `json:"worst_day_loss"`

	// SampleDays is the number of distinct days used.
	SampleDays int `json:"sample_days"`
	// WindowDays is the lookback window in calendar days.
	WindowDays int `json:"window_days"`

	// CapacityMultiplier is a rough estimate of how much more
	// notional the player could hold at similar risk (e.g. 1.8x).
	CapacityMultiplier float64 `json:"capacity_multiplier"`

	// LowSample is true when VaR/ES are unreliable due to small sample size (<20 days).
	LowSample bool `json:"low_sample"`
}

const (
	// portfolioLookbackDays defines how far back we look for wallet transactions.
	portfolioLookbackDays = 180
	// minRiskSampleDays is the minimal number of distinct days required.
	// Keep this low so that new characters still see a rough estimate.
	minRiskSampleDays = 5
)

// ComputePortfolioRiskFromTransactions builds a simple daily PnL series from wallet
// transactions and estimates risk metrics. It intentionally stays heuristic, since
// EVE trading flows are noisy but we only need a coarse, intuitive signal.
func ComputePortfolioRiskFromTransactions(txns []esi.WalletTransaction) *PortfolioRiskSummary {
	if len(txns) == 0 {
		return nil
	}

	// Aggregate daily PnL over the lookback window.
	now := time.Now().UTC()
	cutoff := now.AddDate(0, 0, -portfolioLookbackDays)

	dailyPnL := make(map[time.Time]float64)
	for _, tx := range txns {
		t, err := time.Parse(time.RFC3339, tx.Date)
		if err != nil {
			// Skip badly formatted dates.
			continue
		}
		if t.Before(cutoff) {
			continue
		}
		day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		amount := tx.UnitPrice * float64(tx.Quantity)
		if tx.IsBuy {
			amount = -amount
		}
		dailyPnL[day] += amount
	}

	if len(dailyPnL) < minRiskSampleDays {
		// Not enough data for a meaningful estimate.
		return nil
	}

	// Turn map into sorted slice.
	type dayPnl struct {
		day time.Time
		pnl float64
	}
	var series []dayPnl
	for d, pnl := range dailyPnL {
		series = append(series, dayPnl{day: d, pnl: pnl})
	}
	sort.Slice(series, func(i, j int) bool {
		return series[i].day.Before(series[j].day)
	})

	// Extract PnL values.
	pnls := make([]float64, len(series))
	for i, dp := range series {
		pnls[i] = dp.pnl
	}

	typical := robustScale(pnls)
	if typical <= 0 {
		return nil
	}

	// Convert to "returns" relative to typical daily scale.
	returns := make([]float64, len(pnls))
	for i, v := range pnls {
		returns[i] = v / typical
	}

	// Compute empirical VaR/ES on the PnL distribution directly.
	var95, var99, es95, es99 := portfolioVarEs(pnls)
	worstLoss := minFloat64(pnls)

	// Volatility proxy on normalized returns.
	std := math.Sqrt(variance(returns))

	// Simple risk score: map std to 0-100 with soft cap.
	// For EVE daily flows, std ~1 is "balanced", ~2+ is "high".
	score := std * 40 // std=1 -> 40, std=2 -> 80
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	level := "balanced"
	switch {
	case score < 30:
		level = "safe"
	case score > 70:
		level = "high"
	}

	// Capacity multiplier: if Sharpe-like ratio is ok and score not extreme,
	// we tell the player they can scale up a bit.
	meanRet := mean(returns)
	capacity := 1.0
	if score < 70 && std > 0 {
		sharpeLike := meanRet / std
		switch {
		case sharpeLike > 1.0:
			capacity = 2.0
		case sharpeLike > 0.5:
			capacity = 1.5
		default:
			capacity = 1.2
		}
	}

	return &PortfolioRiskSummary{
		RiskScore:          score,
		RiskLevel:          level,
		Var95:              -var95, // report as positive loss
		Var99:              -var99,
		ES95:               -es95,
		ES99:               -es99,
		TypicalDailyPnl:    typical,
		WorstDayLoss:       -worstLoss,
		SampleDays:         len(pnls),
		WindowDays:         portfolioLookbackDays,
		CapacityMultiplier: capacity,
		LowSample:         len(pnls) < 20, // VaR/ES unreliable with <20 data points
	}
}

func robustScale(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	absVals := make([]float64, 0, len(x))
	for _, v := range x {
		absVals = append(absVals, math.Abs(v))
	}
	sort.Float64s(absVals)
	// Use median of |PnL| as typical daily scale.
	n := len(absVals)
	if n%2 == 1 {
		return absVals[n/2]
	}
	return 0.5 * (absVals[n/2-1] + absVals[n/2])
}

func portfolioVarEs(pnls []float64) (var95, var99, es95, es99 float64) {
	if len(pnls) == 0 {
		return
	}
	sorted := make([]float64, len(pnls))
	copy(sorted, pnls)
	sort.Float64s(sorted) // ascending: biggest loss first (most negative)

	n := len(sorted)
	idx95 := int(math.Floor(0.05 * float64(n)))
	idx99 := int(math.Floor(0.01 * float64(n)))
	if idx95 < 0 {
		idx95 = 0
	}
	if idx95 >= n {
		idx95 = n - 1
	}
	if idx99 < 0 {
		idx99 = 0
	}
	if idx99 >= n {
		idx99 = n - 1
	}

	var95 = sorted[idx95]
	var99 = sorted[idx99]

	// ES: average of worst X% days (most negative values).
	cut95 := idx95 + 1
	if cut95 < 1 {
		cut95 = 1
	}
	es95 = mean(sorted[:cut95])

	cut99 := idx99 + 1
	if cut99 < 1 {
		cut99 = 1
	}
	es99 = mean(sorted[:cut99])
	return
}

func minFloat64(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	m := x[0]
	for _, v := range x[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func mean(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	var sum float64
	for _, v := range x {
		sum += v
	}
	return sum / float64(len(x))
}

