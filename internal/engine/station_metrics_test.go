package engine

import (
	"math"
	"testing"
	"time"

	"eve-flipper/internal/esi"
)

// --- CalcVWAP: VWAP = sum(avg*vol) / sum(vol) ---

func TestCalcVWAP_Exact(t *testing.T) {
	base := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	d1 := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	d2 := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	history := []esi.HistoryEntry{
		{Date: base, Average: 100, Highest: 105, Lowest: 95, Volume: 1000, OrderCount: 10},
		{Date: d1, Average: 200, Highest: 210, Lowest: 190, Volume: 2000, OrderCount: 20},
		{Date: d2, Average: 150, Highest: 155, Lowest: 145, Volume: 500, OrderCount: 5},
	}
	// VWAP = (100*1000 + 200*2000 + 150*500) / (1000+2000+500) = (100000+400000+75000)/3500 = 575000/3500 = 164.285714...
	got := CalcVWAP(history, 30)
	want := 100.0*1000 + 200.0*2000 + 150.0*500
	want /= 1000 + 2000 + 500
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("CalcVWAP = %v, want %v", got, want)
	}
}

func TestCalcVWAP_EmptyAndZeroVolume(t *testing.T) {
	if got := CalcVWAP(nil, 7); got != 0 {
		t.Errorf("CalcVWAP(nil) = %v, want 0", got)
	}
	history := []esi.HistoryEntry{{Date: time.Now().Format("2006-01-02"), Average: 100, Volume: 0}}
	if got := CalcVWAP(history, 7); got != 0 {
		t.Errorf("CalcVWAP(zero volume) = %v, want 0", got)
	}
}

// --- CalcPVI: stdDev of daily range % = (Highest-Lowest)/Average*100 ---

func TestCalcPVI_Exact(t *testing.T) {
	base := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	d1 := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	// Day1: range% = (110-90)/100*100 = 20. Day2: (120-80)/100*100 = 40. Mean=30, variance = ((20-30)^2+(40-30)^2)/2 = 100, std = 10.
	history := []esi.HistoryEntry{
		{Date: base, Average: 100, Highest: 110, Lowest: 90, Volume: 1000},
		{Date: d1, Average: 100, Highest: 120, Lowest: 80, Volume: 1000},
	}
	got := CalcPVI(history, 30)
	want := 10.0 // std of 20 and 40
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("CalcPVI = %v, want %v", got, want)
	}
}

func TestCalcPVI_LessThanTwoDays(t *testing.T) {
	history := []esi.HistoryEntry{{Date: time.Now().Format("2006-01-02"), Average: 100, Highest: 105, Lowest: 95, Volume: 100}}
	if got := CalcPVI(history, 7); got != 0 {
		t.Errorf("CalcPVI(1 day) = %v, want 0", got)
	}
}

// --- CalcOBDS: min(buyDepth, sellDepth) / capitalRequired, where depth is volume within ±5% of best ---

func TestCalcOBDS_Exact(t *testing.T) {
	// Best buy = 100, best sell = 110. Buy orders within 5% of 100: 100 to 100 (only 100). Sell within 5% of 110: 110 to 115.5.
	// sumVolumeWithinPercent: for buy orders, priceDiff = (refPrice - o.Price)/refPrice*100, count if 0<=priceDiff<=5.
	// So buy at 100: (100-100)/100*100=0 -> count. Buy at 98: (100-98)/100*100=2 -> count. Buy at 94: 6 -> no.
	// For sell: priceDiff = (o.Price - refPrice)/refPrice*100. Sell at 110: 0. Sell at 112: 1.82. Sell at 115: 4.54. All within 5%.
	buyOrders := []esi.MarketOrder{
		{Price: 100, VolumeRemain: 100},
		{Price: 98, VolumeRemain: 50},
	}
	sellOrders := []esi.MarketOrder{
		{Price: 110, VolumeRemain: 200},
		{Price: 112, VolumeRemain: 100},
	}
	// Best buy = 100, best sell = 110.
	// Buy depth within 5%: 100@100 (0%), 98@50 (2%) -> 100*100 + 50*98 = 10000+4900 = 14900.
	// Sell depth within 5%: 110@200 (0%), 112@100 (1.82%) -> 200*110 + 100*112 = 22000+11200 = 33200.
	// minDepth = 14900, capitalRequired = 10000 -> OBDS = 14900/10000 = 1.49
	got := CalcOBDS(buyOrders, sellOrders, 10000)
	want := 14900.0 / 10000.0
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("CalcOBDS = %v, want %v", got, want)
	}
}

func TestCalcOBDS_ZeroCapitalOrEmpty(t *testing.T) {
	orders := []esi.MarketOrder{{Price: 100, VolumeRemain: 10}}
	if got := CalcOBDS(orders, nil, 1000); got != 0 {
		t.Errorf("CalcOBDS(nil sell) = %v, want 0", got)
	}
	if got := CalcOBDS(nil, orders, 1000); got != 0 {
		t.Errorf("CalcOBDS(nil buy) = %v, want 0", got)
	}
	if got := CalcOBDS(orders, orders, 0); got != 0 {
		t.Errorf("CalcOBDS(capital 0) = %v, want 0", got)
	}
}

// --- CalcSDS: scam score 0-100 (best buy < 50% VWAP +30, volume mismatch +25, single order dominance +25, no recent trades +20) ---

func TestCalcSDS_NoOrders(t *testing.T) {
	got := CalcSDS(nil, nil, 100)
	if got != 100 {
		t.Errorf("CalcSDS(no buy orders) = %v, want 100 (suspicious)", got)
	}
}

func TestCalcSDS_BestBuyBelowHalfVWAP(t *testing.T) {
	// VWAP=100, best buy 40 -> 40 < 50% VWAP -> +30 (plus possible +25 dominance, +20 no recent trades if history nil)
	orders := []esi.MarketOrder{{Price: 40, VolumeRemain: 100}}
	got := CalcSDS(orders, nil, 100)
	if got < 30 {
		t.Errorf("CalcSDS(best buy 40, VWAP 100) = %v, want >= 30", got)
	}
}

func TestCalcSDS_BestBuyAboveHalfVWAP_NoOtherTriggers(t *testing.T) {
	// Best buy 60 >= 50% VWAP; use recent history and two orders so no +25 dominance, no +20 no recent trades
	today := time.Now().Format("2006-01-02")
	history := []esi.HistoryEntry{{Date: today, Average: 100, Volume: 1000}}
	orders := []esi.MarketOrder{{Price: 60, VolumeRemain: 50}, {Price: 58, VolumeRemain: 50}}
	got := CalcSDS(orders, history, 100)
	if got != 0 {
		t.Errorf("CalcSDS(best buy 60, recent history, two orders) = %v, want 0", got)
	}
}

// --- normalize and CalcCTS (composite) ---

func TestNormalize(t *testing.T) {
	// normalize(value, minVal, maxVal) -> clamp (value-min)/(max-min) to [0,1]
	if got := normalize(50, 0, 100); math.Abs(got-0.5) > 1e-9 {
		t.Errorf("normalize(50,0,100) = %v, want 0.5", got)
	}
	if got := normalize(-10, 0, 100); got != 0 {
		t.Errorf("normalize(-10,0,100) = %v, want 0", got)
	}
	if got := normalize(150, 0, 100); got != 1 {
		t.Errorf("normalize(150,0,100) = %v, want 1", got)
	}
}

func TestCalcCTS_Bounds(t *testing.T) {
	// CTS is weighted sum of normalized components; result should be in reasonable range (not strictly 0-100 due to formula)
	got := CalcCTS(50, 1, 10, 5, 10, 50)
	if got < 0 || got > 100 {
		t.Errorf("CalcCTS should be in [0,100] range, got %v", got)
	}
}
