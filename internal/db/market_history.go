package db

import (
	"log"
	"time"

	"eve-flipper/internal/esi"
)

// GetHistory retrieves cached market history for a region/type pair.
// Returns nil, false if not cached or if cache is older than 24 hours.
func (d *DB) GetMarketHistory(regionID int32, typeID int32) ([]esi.HistoryEntry, bool) {
	var updatedAt string
	err := d.sql.QueryRow(
		"SELECT updated_at FROM market_history_meta WHERE region_id=? AND type_id=?",
		regionID, typeID,
	).Scan(&updatedAt)
	if err != nil {
		return nil, false
	}

	// Check if cache is fresh (< 24 hours)
	t, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil || time.Since(t) > 24*time.Hour {
		return nil, false
	}

	rows, err := d.sql.Query(
		"SELECT date, average, highest, lowest, volume, order_count FROM market_history WHERE region_id=? AND type_id=? ORDER BY date",
		regionID, typeID,
	)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var entries []esi.HistoryEntry
	for rows.Next() {
		var e esi.HistoryEntry
		if err := rows.Scan(&e.Date, &e.Average, &e.Highest, &e.Lowest, &e.Volume, &e.OrderCount); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	if len(entries) == 0 {
		return nil, false
	}
	return entries, true
}

// SetMarketHistory stores market history entries in the cache.
// Only entries from the last 90 days are stored to bound database growth.
func (d *DB) SetMarketHistory(regionID int32, typeID int32, entries []esi.HistoryEntry) {
	tx, err := d.sql.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	// Delete old entries for this region+type
	tx.Exec("DELETE FROM market_history WHERE region_id=? AND type_id=?", regionID, typeID)

	stmt, err := tx.Prepare("INSERT INTO market_history (region_id, type_id, date, average, highest, lowest, volume, order_count) VALUES (?,?,?,?,?,?,?,?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	// Only keep last 90 days of data
	cutoff := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	for _, e := range entries {
		if e.Date >= cutoff {
			stmt.Exec(regionID, typeID, e.Date, e.Average, e.Highest, e.Lowest, e.Volume, e.OrderCount)
		}
	}

	// Update meta
	tx.Exec(
		"INSERT OR REPLACE INTO market_history_meta (region_id, type_id, updated_at) VALUES (?,?,?)",
		regionID, typeID, time.Now().UTC().Format(time.RFC3339),
	)

	tx.Commit()
}

// CleanupOldHistory removes market history data older than 90 days and
// meta entries that haven't been refreshed in over 30 days.
// Should be called periodically (e.g. on startup or daily) to prevent
// unbounded SQLite database growth.
func (d *DB) CleanupOldHistory() {
	cutoffDate := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	cutoffMeta := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)

	// Delete history rows older than 90 days
	res, err := d.sql.Exec("DELETE FROM market_history WHERE date < ?", cutoffDate)
	if err != nil {
		log.Printf("[DB] CleanupOldHistory: history delete error: %v", err)
	} else if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("[DB] CleanupOldHistory: removed %d old history rows", n)
	}

	// Delete meta entries not refreshed in 30 days (stale type+region pairs)
	res, err = d.sql.Exec("DELETE FROM market_history_meta WHERE updated_at < ?", cutoffMeta)
	if err != nil {
		log.Printf("[DB] CleanupOldHistory: meta delete error: %v", err)
	} else if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("[DB] CleanupOldHistory: removed %d stale meta entries", n)
	}

	// Delete orphaned history rows (meta was removed but history rows remain)
	res, err = d.sql.Exec(`
		DELETE FROM market_history
		WHERE (region_id, type_id) NOT IN (
			SELECT region_id, type_id FROM market_history_meta
		)
	`)
	if err != nil {
		log.Printf("[DB] CleanupOldHistory: orphan delete error: %v", err)
	} else if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("[DB] CleanupOldHistory: removed %d orphaned history rows", n)
	}
}
