package engine

import (
	"math"
	"sort"
	"time"

	"eve-flipper/internal/esi"
)

// PortfolioPnL is the full P&L analytics response for the character popup.
type PortfolioPnL struct {
	DailyPnL    []DailyPnLEntry   `json:"daily_pnl"`
	Summary     PortfolioPnLStats `json:"summary"`
	TopItems    []ItemPnL         `json:"top_items"`
	TopStations []StationPnL      `json:"top_stations"`
}

// DailyPnLEntry represents one day's trading activity.
type DailyPnLEntry struct {
	Date          string  `json:"date"` // YYYY-MM-DD
	BuyTotal      float64 `json:"buy_total"`
	SellTotal     float64 `json:"sell_total"`
	NetPnL        float64 `json:"net_pnl"`
	CumulativePnL float64 `json:"cumulative_pnl"`
	DrawdownPct   float64 `json:"drawdown_pct"` // drawdown from cumulative peak (0 to -100)
	Transactions  int     `json:"transactions"`
}

// PortfolioPnLStats is the aggregated summary across the period.
type PortfolioPnLStats struct {
	TotalPnL       float64 `json:"total_pnl"`
	AvgDailyPnL    float64 `json:"avg_daily_pnl"`
	BestDayPnL     float64 `json:"best_day_pnl"`
	BestDayDate    string  `json:"best_day_date"`
	WorstDayPnL    float64 `json:"worst_day_pnl"`
	WorstDayDate   string  `json:"worst_day_date"`
	ProfitableDays int     `json:"profitable_days"`
	LosingDays     int     `json:"losing_days"`
	TotalDays      int     `json:"total_days"`
	WinRate        float64 `json:"win_rate"` // 0-100%
	TotalBought    float64 `json:"total_bought"`
	TotalSold      float64 `json:"total_sold"`
	ROIPercent     float64 `json:"roi_percent"`

	// Enhanced analytics
	SharpeRatio        float64 `json:"sharpe_ratio"`         // annualized: mean/std * sqrt(365)
	MaxDrawdownPct     float64 `json:"max_drawdown_pct"`     // deepest cumulative drawdown %
	MaxDrawdownISK     float64 `json:"max_drawdown_isk"`     // deepest drawdown in ISK
	MaxDrawdownDays    int     `json:"max_drawdown_days"`    // duration from peak to trough
	CalmarRatio        float64 `json:"calmar_ratio"`         // annualized return / max drawdown
	ProfitFactor       float64 `json:"profit_factor"`        // gross profit / gross loss
	AvgWin             float64 `json:"avg_win"`              // average winning day ISK
	AvgLoss            float64 `json:"avg_loss"`             // average losing day ISK
	ExpectancyPerTrade float64 `json:"expectancy_per_trade"` // (win_rate * avg_win) - (loss_rate * avg_loss)
}

// StationPnL is a per-station breakdown of trading activity.
type StationPnL struct {
	LocationID   int64   `json:"location_id"`
	LocationName string  `json:"location_name"`
	TotalBought  float64 `json:"total_bought"`
	TotalSold    float64 `json:"total_sold"`
	NetPnL       float64 `json:"net_pnl"`
	Transactions int     `json:"transactions"`
}

// ItemPnL is the per-item breakdown of trading activity.
type ItemPnL struct {
	TypeID        int32   `json:"type_id"`
	TypeName      string  `json:"type_name"`
	TotalBought   float64 `json:"total_bought"`
	TotalSold     float64 `json:"total_sold"`
	NetPnL        float64 `json:"net_pnl"`
	QtyBought     int64   `json:"qty_bought"`
	QtySold       int64   `json:"qty_sold"`
	AvgBuyPrice   float64 `json:"avg_buy_price"`
	AvgSellPrice  float64 `json:"avg_sell_price"`
	MarginPercent float64 `json:"margin_percent"`
	Transactions  int     `json:"transactions"`
}

// ComputePortfolioPnL builds a full P&L analysis from wallet transactions.
// lookbackDays controls how far back to look (e.g. 7, 30, 90, 180).
func ComputePortfolioPnL(txns []esi.WalletTransaction, lookbackDays int) *PortfolioPnL {
	if len(txns) == 0 {
		return &PortfolioPnL{
			DailyPnL:    []DailyPnLEntry{},
			TopItems:    []ItemPnL{},
			TopStations: []StationPnL{},
		}
	}

	now := time.Now().UTC()
	cutoff := now.AddDate(0, 0, -lookbackDays)

	// Aggregate daily PnL, per-item stats, and per-station stats.
	type dayKey string
	dayMap := make(map[dayKey]*DailyPnLEntry)
	itemMap := make(map[int32]*ItemPnL)
	stationMap := make(map[int64]*StationPnL)

	for _, tx := range txns {
		t, err := time.Parse(time.RFC3339, tx.Date)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			continue
		}

		// Daily aggregation
		dk := dayKey(t.Format("2006-01-02"))
		entry, ok := dayMap[dk]
		if !ok {
			entry = &DailyPnLEntry{Date: string(dk)}
			dayMap[dk] = entry
		}

		amount := tx.UnitPrice * float64(tx.Quantity)
		entry.Transactions++
		if tx.IsBuy {
			entry.BuyTotal += amount
		} else {
			entry.SellTotal += amount
		}

		// Per-item aggregation
		item, ok := itemMap[tx.TypeID]
		if !ok {
			item = &ItemPnL{
				TypeID:   tx.TypeID,
				TypeName: tx.TypeName,
			}
			itemMap[tx.TypeID] = item
		}
		item.Transactions++
		if tx.IsBuy {
			item.TotalBought += amount
			item.QtyBought += int64(tx.Quantity)
		} else {
			item.TotalSold += amount
			item.QtySold += int64(tx.Quantity)
		}

		// Per-station aggregation
		st, ok := stationMap[tx.LocationID]
		if !ok {
			st = &StationPnL{
				LocationID:   tx.LocationID,
				LocationName: tx.LocationName,
			}
			stationMap[tx.LocationID] = st
		}
		st.Transactions++
		if tx.IsBuy {
			st.TotalBought += amount
		} else {
			st.TotalSold += amount
		}
	}

	// Sort days chronologically.
	days := make([]DailyPnLEntry, 0, len(dayMap))
	for _, entry := range dayMap {
		entry.NetPnL = entry.SellTotal - entry.BuyTotal
		days = append(days, *entry)
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

	// Cumulative PnL and drawdown from peak.
	cumulative := 0.0
	cumulativePeak := 0.0
	maxDrawdownISK := 0.0
	maxDrawdownPeakIdx := 0
	maxDrawdownTroughIdx := 0
	currentPeakIdx := 0

	for i := range days {
		cumulative += days[i].NetPnL
		days[i].CumulativePnL = cumulative

		if cumulative > cumulativePeak {
			cumulativePeak = cumulative
			currentPeakIdx = i
		}

		drawdownISK := cumulative - cumulativePeak
		if cumulativePeak > 0 {
			days[i].DrawdownPct = drawdownISK / cumulativePeak * 100
		}

		if drawdownISK < maxDrawdownISK {
			maxDrawdownISK = drawdownISK
			maxDrawdownPeakIdx = currentPeakIdx
			maxDrawdownTroughIdx = i
		}
	}

	// Summary stats
	summary := PortfolioPnLStats{
		TotalDays: len(days),
	}

	if len(days) > 0 {
		summary.BestDayPnL = days[0].NetPnL
		summary.BestDayDate = days[0].Date
		summary.WorstDayPnL = days[0].NetPnL
		summary.WorstDayDate = days[0].Date
	}

	var grossProfit, grossLoss float64
	var totalWinISK, totalLossISK float64

	for _, d := range days {
		summary.TotalPnL += d.NetPnL
		summary.TotalBought += d.BuyTotal
		summary.TotalSold += d.SellTotal

		if d.NetPnL > 0 {
			summary.ProfitableDays++
			grossProfit += d.NetPnL
			totalWinISK += d.NetPnL
		} else if d.NetPnL < 0 {
			summary.LosingDays++
			grossLoss += -d.NetPnL // store as positive
			totalLossISK += -d.NetPnL
		}

		if d.NetPnL > summary.BestDayPnL {
			summary.BestDayPnL = d.NetPnL
			summary.BestDayDate = d.Date
		}
		if d.NetPnL < summary.WorstDayPnL {
			summary.WorstDayPnL = d.NetPnL
			summary.WorstDayDate = d.Date
		}
	}

	if summary.TotalDays > 0 {
		summary.AvgDailyPnL = summary.TotalPnL / float64(summary.TotalDays)
		summary.WinRate = float64(summary.ProfitableDays) / float64(summary.TotalDays) * 100
	}
	// ROI: use time-weighted average capital deployed instead of gross purchases.
	// The naive PnL/GrossPurchases systematically underestimates ROI for high-turnover
	// traders because GrossPurchases counts the same ISK multiple times across
	// buy-sell cycles (e.g., 100M ISK turned over 10× looks like 1B invested).
	if len(days) > 0 {
		var cumBuy, cumSell, capitalSum float64
		for _, d := range days {
			cumBuy += d.BuyTotal
			cumSell += d.SellTotal
			deployed := cumBuy - cumSell
			if deployed > 0 {
				capitalSum += deployed
			}
		}
		avgCapital := capitalSum / float64(len(days))
		if avgCapital > 0 {
			summary.ROIPercent = summary.TotalPnL / avgCapital * 100
		} else if summary.TotalBought > 0 {
			// Fallback for edge cases (e.g., all inventory from before the period).
			summary.ROIPercent = summary.TotalPnL / summary.TotalBought * 100
		}
	}

	// Sharpe ratio: annualized = (mean daily PnL / std daily PnL) * sqrt(365)
	if summary.TotalDays >= 2 {
		dailyPnLs := make([]float64, len(days))
		for i, d := range days {
			dailyPnLs[i] = d.NetPnL
		}
		mu := mean(dailyPnLs)
		sigma := math.Sqrt(variance(dailyPnLs))
		if sigma > 0 {
			summary.SharpeRatio = (mu / sigma) * math.Sqrt(365)
		}
	}

	// Max drawdown
	summary.MaxDrawdownISK = -maxDrawdownISK // report as positive
	if cumulativePeak > 0 {
		summary.MaxDrawdownPct = -maxDrawdownISK / cumulativePeak * 100
	}
	// MaxDrawdownDays: use actual calendar days (not array index positions)
	// so weekends and inactive days are properly counted in the duration.
	if maxDrawdownTroughIdx > maxDrawdownPeakIdx {
		peakDate, errP := time.Parse("2006-01-02", days[maxDrawdownPeakIdx].Date)
		troughDate, errT := time.Parse("2006-01-02", days[maxDrawdownTroughIdx].Date)
		if errP == nil && errT == nil {
			summary.MaxDrawdownDays = int(troughDate.Sub(peakDate).Hours() / 24)
		} else {
			// Fallback to index difference if date parsing fails.
			summary.MaxDrawdownDays = maxDrawdownTroughIdx - maxDrawdownPeakIdx
		}
	}

	// Calmar ratio: annualized return / max drawdown
	if summary.MaxDrawdownISK > 0 && summary.TotalDays > 0 {
		annualizedReturn := summary.TotalPnL * 365 / float64(summary.TotalDays)
		summary.CalmarRatio = annualizedReturn / summary.MaxDrawdownISK
	}

	// Profit factor: gross profit / gross loss
	if grossLoss > 0 {
		summary.ProfitFactor = grossProfit / grossLoss
	}

	// Average win/loss
	if summary.ProfitableDays > 0 {
		summary.AvgWin = totalWinISK / float64(summary.ProfitableDays)
	}
	if summary.LosingDays > 0 {
		summary.AvgLoss = totalLossISK / float64(summary.LosingDays)
	}

	// Expectancy per trade: (win_rate * avg_win) - (loss_rate * avg_loss)
	if summary.TotalDays > 0 {
		winRate := float64(summary.ProfitableDays) / float64(summary.TotalDays)
		lossRate := float64(summary.LosingDays) / float64(summary.TotalDays)
		summary.ExpectancyPerTrade = winRate*summary.AvgWin - lossRate*summary.AvgLoss
	}

	// Per-item stats: compute averages and margin.
	items := make([]ItemPnL, 0, len(itemMap))
	for _, item := range itemMap {
		item.NetPnL = item.TotalSold - item.TotalBought
		if item.QtyBought > 0 {
			item.AvgBuyPrice = item.TotalBought / float64(item.QtyBought)
		}
		if item.QtySold > 0 {
			item.AvgSellPrice = item.TotalSold / float64(item.QtySold)
		}
		if item.AvgBuyPrice > 0 && item.AvgSellPrice > 0 {
			item.MarginPercent = (item.AvgSellPrice - item.AvgBuyPrice) / item.AvgBuyPrice * 100
		}
		items = append(items, *item)
	}

	// Sort items by absolute PnL (most impactful first).
	sort.Slice(items, func(i, j int) bool {
		absI := items[i].NetPnL
		if absI < 0 {
			absI = -absI
		}
		absJ := items[j].NetPnL
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})

	// Limit to top 50 items.
	if len(items) > 50 {
		items = items[:50]
	}

	// Per-station stats
	stations := make([]StationPnL, 0, len(stationMap))
	for _, st := range stationMap {
		st.NetPnL = st.TotalSold - st.TotalBought
		stations = append(stations, *st)
	}
	sort.Slice(stations, func(i, j int) bool {
		absI := stations[i].NetPnL
		if absI < 0 {
			absI = -absI
		}
		absJ := stations[j].NetPnL
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})
	if len(stations) > 20 {
		stations = stations[:20]
	}

	return &PortfolioPnL{
		DailyPnL:    days,
		Summary:     summary,
		TopItems:    items,
		TopStations: stations,
	}
}
