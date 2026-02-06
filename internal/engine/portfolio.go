package engine

import (
	"sort"
	"time"

	"eve-flipper/internal/esi"
)

// PortfolioPnL is the full P&L analytics response for the character popup.
type PortfolioPnL struct {
	DailyPnL []DailyPnLEntry   `json:"daily_pnl"`
	Summary  PortfolioPnLStats `json:"summary"`
	TopItems []ItemPnL         `json:"top_items"`
}

// DailyPnLEntry represents one day's trading activity.
type DailyPnLEntry struct {
	Date          string  `json:"date"` // YYYY-MM-DD
	BuyTotal      float64 `json:"buy_total"`
	SellTotal     float64 `json:"sell_total"`
	NetPnL        float64 `json:"net_pnl"`
	CumulativePnL float64 `json:"cumulative_pnl"`
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
			DailyPnL: []DailyPnLEntry{},
			TopItems: []ItemPnL{},
		}
	}

	now := time.Now().UTC()
	cutoff := now.AddDate(0, 0, -lookbackDays)

	// Aggregate daily PnL and per-item stats.
	type dayKey string
	dayMap := make(map[dayKey]*DailyPnLEntry)
	itemMap := make(map[int32]*ItemPnL)

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

	// Cumulative PnL
	cumulative := 0.0
	for i := range days {
		cumulative += days[i].NetPnL
		days[i].CumulativePnL = cumulative
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

	for _, d := range days {
		summary.TotalPnL += d.NetPnL
		summary.TotalBought += d.BuyTotal
		summary.TotalSold += d.SellTotal

		if d.NetPnL > 0 {
			summary.ProfitableDays++
		} else if d.NetPnL < 0 {
			summary.LosingDays++
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
	if summary.TotalBought > 0 {
		summary.ROIPercent = (summary.TotalSold - summary.TotalBought) / summary.TotalBought * 100
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

	return &PortfolioPnL{
		DailyPnL: days,
		Summary:  summary,
		TopItems: items,
	}
}
