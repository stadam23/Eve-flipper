package db

import (
	"eve-flipper/internal/config"
)

// GetWatchlist returns all watchlist items.
func (d *DB) GetWatchlist() []config.WatchlistItem {
	rows, err := d.sql.Query("SELECT type_id, type_name, added_at, alert_min_margin FROM watchlist ORDER BY added_at DESC")
	if err != nil {
		return []config.WatchlistItem{}
	}
	defer rows.Close()

	var items []config.WatchlistItem
	for rows.Next() {
		var item config.WatchlistItem
		rows.Scan(&item.TypeID, &item.TypeName, &item.AddedAt, &item.AlertMinMargin)
		items = append(items, item)
	}
	if items == nil {
		return []config.WatchlistItem{}
	}
	return items
}

// HasWatchlistItem checks if an item is already in the watchlist.
func (d *DB) HasWatchlistItem(typeID int32) bool {
	var count int
	d.sql.QueryRow("SELECT COUNT(*) FROM watchlist WHERE type_id = ?", typeID).Scan(&count)
	return count > 0
}

// AddWatchlistItem inserts a watchlist item. Returns true if inserted, false if duplicate.
func (d *DB) AddWatchlistItem(item config.WatchlistItem) bool {
	res, err := d.sql.Exec(
		"INSERT OR IGNORE INTO watchlist (type_id, type_name, added_at, alert_min_margin) VALUES (?, ?, ?, ?)",
		item.TypeID, item.TypeName, item.AddedAt, item.AlertMinMargin,
	)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

// DeleteWatchlistItem removes a watchlist item by type ID.
func (d *DB) DeleteWatchlistItem(typeID int32) {
	d.sql.Exec("DELETE FROM watchlist WHERE type_id = ?", typeID)
}

// UpdateWatchlistItem updates the alert threshold for a watchlist item.
func (d *DB) UpdateWatchlistItem(typeID int32, alertMinMargin float64) {
	d.sql.Exec("UPDATE watchlist SET alert_min_margin = ? WHERE type_id = ?", alertMinMargin, typeID)
}
