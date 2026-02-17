package db

import (
	"eve-flipper/internal/config"
)

// GetWatchlist returns all watchlist items.
func (d *DB) GetWatchlist() []config.WatchlistItem {
	return d.GetWatchlistForUser(DefaultUserID)
}

// GetWatchlistForUser returns all watchlist items for a specific user.
func (d *DB) GetWatchlistForUser(userID string) []config.WatchlistItem {
	userID = normalizeUserID(userID)

	rows, err := d.sql.Query(`
		SELECT type_id, type_name, added_at, alert_min_margin, alert_enabled, alert_metric, alert_threshold
		  FROM watchlist
		 WHERE user_id = ?
		 ORDER BY added_at DESC
	`, userID)
	if err != nil {
		return []config.WatchlistItem{}
	}
	defer rows.Close()

	var items []config.WatchlistItem
	for rows.Next() {
		var item config.WatchlistItem
		rows.Scan(
			&item.TypeID,
			&item.TypeName,
			&item.AddedAt,
			&item.AlertMinMargin,
			&item.AlertEnabled,
			&item.AlertMetric,
			&item.AlertThreshold,
		)
		if item.AlertMetric == "" {
			item.AlertMetric = "margin_percent"
		}
		if item.AlertThreshold <= 0 && item.AlertMinMargin > 0 {
			item.AlertThreshold = item.AlertMinMargin
		}
		items = append(items, item)
	}
	if items == nil {
		return []config.WatchlistItem{}
	}
	return items
}

// HasWatchlistItem checks if an item is already in the watchlist.
func (d *DB) HasWatchlistItem(typeID int32) bool {
	return d.HasWatchlistItemForUser(DefaultUserID, typeID)
}

// HasWatchlistItemForUser checks if an item is already in the watchlist for a specific user.
func (d *DB) HasWatchlistItemForUser(userID string, typeID int32) bool {
	userID = normalizeUserID(userID)

	var count int
	d.sql.QueryRow("SELECT COUNT(*) FROM watchlist WHERE user_id = ? AND type_id = ?", userID, typeID).Scan(&count)
	return count > 0
}

// AddWatchlistItem inserts a watchlist item. Returns true if inserted, false if duplicate.
func (d *DB) AddWatchlistItem(item config.WatchlistItem) bool {
	return d.AddWatchlistItemForUser(DefaultUserID, item)
}

// AddWatchlistItemForUser inserts a watchlist item for a specific user.
// Returns true if inserted, false if duplicate.
func (d *DB) AddWatchlistItemForUser(userID string, item config.WatchlistItem) bool {
	userID = normalizeUserID(userID)

	if item.AlertMetric == "" {
		item.AlertMetric = "margin_percent"
	}
	if item.AlertThreshold <= 0 && item.AlertMinMargin > 0 {
		item.AlertThreshold = item.AlertMinMargin
	}
	if item.AlertThreshold > 0 && !item.AlertEnabled {
		item.AlertEnabled = true
	}
	if item.AlertMetric == "margin_percent" {
		item.AlertMinMargin = item.AlertThreshold
	} else if item.AlertMinMargin < 0 {
		item.AlertMinMargin = 0
	}
	res, err := d.sql.Exec(
		`INSERT OR IGNORE INTO watchlist
		   (user_id, type_id, type_name, added_at, alert_min_margin, alert_enabled, alert_metric, alert_threshold)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID,
		item.TypeID,
		item.TypeName,
		item.AddedAt,
		item.AlertMinMargin,
		item.AlertEnabled,
		item.AlertMetric,
		item.AlertThreshold,
	)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

// DeleteWatchlistItem removes a watchlist item by type ID.
func (d *DB) DeleteWatchlistItem(typeID int32) {
	d.DeleteWatchlistItemForUser(DefaultUserID, typeID)
}

// DeleteWatchlistItemForUser removes a watchlist item by type ID for a specific user.
func (d *DB) DeleteWatchlistItemForUser(userID string, typeID int32) {
	userID = normalizeUserID(userID)
	d.sql.Exec("DELETE FROM watchlist WHERE user_id = ? AND type_id = ?", userID, typeID)
}

// UpdateWatchlistItem updates alert settings for a watchlist item.
func (d *DB) UpdateWatchlistItem(typeID int32, alertMinMargin float64, alertEnabled bool, alertMetric string, alertThreshold float64) {
	d.UpdateWatchlistItemForUser(DefaultUserID, typeID, alertMinMargin, alertEnabled, alertMetric, alertThreshold)
}

// UpdateWatchlistItemForUser updates alert settings for a watchlist item for a specific user.
func (d *DB) UpdateWatchlistItemForUser(userID string, typeID int32, alertMinMargin float64, alertEnabled bool, alertMetric string, alertThreshold float64) {
	userID = normalizeUserID(userID)

	if alertMetric == "" {
		alertMetric = "margin_percent"
	}
	if alertThreshold < 0 {
		alertThreshold = 0
	}
	if alertMetric == "margin_percent" {
		alertMinMargin = alertThreshold
	} else if alertMinMargin < 0 {
		alertMinMargin = 0
	}
	d.sql.Exec(
		`UPDATE watchlist
		    SET alert_min_margin = ?, alert_enabled = ?, alert_metric = ?, alert_threshold = ?
		  WHERE user_id = ? AND type_id = ?`,
		alertMinMargin,
		alertEnabled,
		alertMetric,
		alertThreshold,
		userID,
		typeID,
	)
}
