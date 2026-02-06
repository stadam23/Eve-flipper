package engine

import (
	"math"
	"testing"
	"time"

	"eve-flipper/internal/esi"
)

// --- Pure math helpers: exact expected values ---

func TestMean(t *testing.T) {
	tests := []struct {
		name string
		x    []float64
		want float64
	}{
		{"empty", nil, 0},
		{"single", []float64{42}, 42},
		{"five", []float64{1, 2, 3, 4, 5}, 3},
		{"negative", []float64{-10, -20, -30}, -20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mean(tt.x)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("mean(%v) = %v, want %v", tt.x, got, tt.want)
			}
		})
	}
}

func TestMinFloat64(t *testing.T) {
	tests := []struct {
		name string
		x    []float64
		want float64
	}{
		{"empty", nil, 0},
		{"single", []float64{7}, 7},
		{"positive", []float64{3, 1, 2}, 1},
		{"negative", []float64{-100, -50, -200}, -200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minFloat64(tt.x)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("minFloat64(%v) = %v, want %v", tt.x, got, tt.want)
			}
		})
	}
}

func TestRobustScale(t *testing.T) {
	tests := []struct {
		name string
		x    []float64
		want float64
	}{
		{"empty", nil, 0},
		{"single", []float64{10}, 10},
		{"odd median of abs", []float64{1, 2, 3, 4, 5}, 3},       // abs 1,2,3,4,5 -> median 3
		{"even median of abs", []float64{1, 2, 3, 4, 5, 6}, 3.5}, // abs 1..6 -> (3+4)/2
		{"negative values", []float64{-5, -4, -3, -2, -1}, 3},    // abs 1..5 -> median 3
		{"mixed", []float64{-100, 50, 80, -20, 40}, 50},          // abs 20,40,50,80,100 -> median 50
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := robustScale(tt.x)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("robustScale(%v) = %v, want %v", tt.x, got, tt.want)
			}
		})
	}
}

func TestPortfolioVarEs(t *testing.T) {
	// Sorted ascending = worst (most negative) first.
	// n=10: idx95 = floor(0.05*10)=0, idx99=0. So var95=first, var99=first. cut95=1, cut99=1 -> es95=es99=first value.
	pnls := []float64{-100, -90, -80, -70, -60, -50, -40, -30, -20, -10}
	var95, var99, es95, es99 := portfolioVarEs(pnls)
	if math.Abs(var95-(-100)) > 1e-9 {
		t.Errorf("var95 = %v, want -100", var95)
	}
	if math.Abs(var99-(-100)) > 1e-9 {
		t.Errorf("var99 = %v, want -100", var99)
	}
	if math.Abs(es95-(-100)) > 1e-9 {
		t.Errorf("es95 = %v, want -100", es95)
	}
	if math.Abs(es99-(-100)) > 1e-9 {
		t.Errorf("es99 = %v, want -100", es99)
	}

	// n=20: idx95 = floor(0.05*20)=1, idx99=0. So var95 = sorted[1], var99 = sorted[0]. ES95 = mean(sorted[:2]), ES99 = mean(sorted[:1]).
	pnls2 := make([]float64, 20)
	for i := range pnls2 {
		pnls2[i] = -100 + float64(i)*5 // -100,-95,-90,...,-5
	}
	var95, var99, es95, es99 = portfolioVarEs(pnls2)
	if math.Abs(var95-(-95)) > 1e-9 {
		t.Errorf("var95 (n=20) = %v, want -95", var95)
	}
	if math.Abs(var99-(-100)) > 1e-9 {
		t.Errorf("var99 (n=20) = %v, want -100", var99)
	}
	wantES95 := (-100.0 + -95.0) / 2.0
	if math.Abs(es95-wantES95) > 1e-9 {
		t.Errorf("es95 (n=20) = %v, want %v", es95, wantES95)
	}
	if math.Abs(es99-(-100)) > 1e-9 {
		t.Errorf("es99 (n=20) = %v, want -100", es99)
	}
}

func TestPortfolioVarEs_EmptyAndSmall(t *testing.T) {
	var95, var99, es95, es99 := portfolioVarEs(nil)
	if var95 != 0 || var99 != 0 || es95 != 0 || es99 != 0 {
		t.Errorf("portfolioVarEs(nil) should return zeros, got var95=%v var99=%v es95=%v es99=%v", var95, var99, es95, es99)
	}

	// Single element: idx95=0, idx99=0, cut95=1, cut99=1
	var95, var99, es95, es99 = portfolioVarEs([]float64{-50})
	if math.Abs(var95-(-50)) > 1e-9 || math.Abs(es95-(-50)) > 1e-9 {
		t.Errorf("portfolioVarEs([-50]) = var95 %v es95 %v, want -50", var95, es95)
	}
}

func TestComputePortfolioRiskFromTransactions_EmptyAndTooFewDays(t *testing.T) {
	if got := ComputePortfolioRiskFromTransactions(nil); got != nil {
		t.Errorf("ComputePortfolioRiskFromTransactions(nil) want nil, got %+v", got)
	}

	// One day only -> less than minRiskSampleDays (5)
	now := time.Now().UTC()
	txns := []esi.WalletTransaction{
		{Date: now.AddDate(0, 0, -1).Format(time.RFC3339), UnitPrice: 100, Quantity: 1, IsBuy: false},
	}
	if got := ComputePortfolioRiskFromTransactions(txns); got != nil {
		t.Errorf("ComputePortfolioRiskFromTransactions(1 day) want nil, got %+v", got)
	}
}

func TestComputePortfolioRiskFromTransactions_EnoughDays_DeterministicOutput(t *testing.T) {
	// Build 10 days of transactions within lookback (180 days). Daily PnL we control:
	// Day -10: +100, -9: -50, -8: +80, -7: -20, -6: +60, -5: +40, -4: -80, -3: +30, -2: +20, -1: +10
	// So daily PnL sum per day: 100, -50, 80, -20, 60, 40, -80, 30, 20, 10.
	// All positive days: sell; negative: buy. One txn per day.
	base := time.Now().UTC().AddDate(0, 0, -10)
	dayPnls := []float64{100, -50, 80, -20, 60, 40, -80, 30, 20, 10}
	var txns []esi.WalletTransaction
	for i, pnl := range dayPnls {
		d := base.AddDate(0, 0, i)
		dateStr := d.Format(time.RFC3339)
		if pnl >= 0 {
			txns = append(txns, esi.WalletTransaction{Date: dateStr, UnitPrice: pnl, Quantity: 1, IsBuy: false})
		} else {
			txns = append(txns, esi.WalletTransaction{Date: dateStr, UnitPrice: -pnl, Quantity: 1, IsBuy: true})
		}
	}

	out := ComputePortfolioRiskFromTransactions(txns)
	if out == nil {
		t.Fatal("ComputePortfolioRiskFromTransactions: expected non-nil summary with 10 days")
	}
	if out.SampleDays != 10 {
		t.Errorf("SampleDays = %d, want 10", out.SampleDays)
	}
	if out.WindowDays != portfolioLookbackDays {
		t.Errorf("WindowDays = %d, want %d", out.WindowDays, portfolioLookbackDays)
	}
	// Worst day loss: min(dayPnls) = -80, we report as positive 80
	if math.Abs(out.WorstDayLoss-80) > 1e-6 {
		t.Errorf("WorstDayLoss = %v, want 80", out.WorstDayLoss)
	}
	// VaR/ES are reported as positive (loss). We have 10 days; 5% of 10 = 0.5 -> idx 0, 1% -> idx 0. So var95/var99 = worst = -80 -> report 80.
	if out.Var95 < 0 || out.Var99 < 0 {
		t.Errorf("Var95/Var99 should be positive (reported loss): Var95=%v Var99=%v", out.Var95, out.Var99)
	}
	// Typical daily PnL = robustScale(dayPnls) = median of abs(100,50,80,20,60,40,80,30,20,10) = median(10,20,20,30,40,50,60,80,80,100) = (40+50)/2 = 45
	wantTypical := 45.0
	if math.Abs(out.TypicalDailyPnl-wantTypical) > 1e-6 {
		t.Errorf("TypicalDailyPnl = %v, want %v", out.TypicalDailyPnl, wantTypical)
	}
	// Risk level: score is std(returns)*40; we only check it's in [0,100] and level is one of safe/balanced/high
	if out.RiskScore < 0 || out.RiskScore > 100 {
		t.Errorf("RiskScore = %v, want in [0,100]", out.RiskScore)
	}
	if out.RiskLevel != "safe" && out.RiskLevel != "balanced" && out.RiskLevel != "high" {
		t.Errorf("RiskLevel = %q, want safe|balanced|high", out.RiskLevel)
	}
}
