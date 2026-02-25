package db

import (
	"encoding/json"
	"time"
)

// ScanRecord represents a scan history entry.
type ScanRecord struct {
	ID          int64           `json:"id"`
	Timestamp   string          `json:"timestamp"`
	Tab         string          `json:"tab"`
	System      string          `json:"system"`
	Count       int             `json:"count"`
	TopProfit   float64         `json:"top_profit"`
	TotalProfit float64         `json:"total_profit"`
	DurationMs  int64           `json:"duration_ms"`
	Params      json.RawMessage `json:"params"`
}

// InsertHistory inserts a scan history record and returns its ID.
func (d *DB) InsertHistory(tab, system string, count int, topProfit float64) int64 {
	result, err := d.sql.Exec(
		"INSERT INTO scan_history (timestamp, tab, system, count, top_profit) VALUES (?, ?, ?, ?, ?)",
		time.Now().Format(time.RFC3339), tab, system, count, topProfit,
	)
	if err != nil {
		return 0
	}
	id, _ := result.LastInsertId()
	return id
}

// InsertHistoryFull inserts a scan history record with all fields.
func (d *DB) InsertHistoryFull(tab, system string, count int, topProfit, totalProfit float64, durationMs int64, params interface{}) int64 {
	paramsJSON, _ := json.Marshal(params)
	result, err := d.sql.Exec(
		"INSERT INTO scan_history (timestamp, tab, system, count, top_profit, total_profit, duration_ms, params_json) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		time.Now().Format(time.RFC3339), tab, system, count, topProfit, totalProfit, durationMs, string(paramsJSON),
	)
	if err != nil {
		return 0
	}
	id, _ := result.LastInsertId()
	return id
}

// GetHistory returns the last N scan history records (newest first).
func (d *DB) GetHistory(limit int) []ScanRecord {
	if limit <= 0 {
		limit = 50
	}
	rows, err := d.sql.Query(
		`SELECT id, timestamp, tab, system, count, top_profit,
		 COALESCE(total_profit, 0), COALESCE(duration_ms, 0), COALESCE(params_json, '{}')
		 FROM scan_history ORDER BY id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return []ScanRecord{}
	}
	defer rows.Close()

	var records []ScanRecord
	for rows.Next() {
		var r ScanRecord
		var paramsStr string
		rows.Scan(&r.ID, &r.Timestamp, &r.Tab, &r.System, &r.Count, &r.TopProfit, &r.TotalProfit, &r.DurationMs, &paramsStr)
		r.Params = json.RawMessage(paramsStr)
		records = append(records, r)
	}
	if records == nil {
		return []ScanRecord{}
	}
	return records
}

// GetHistoryByID returns a single scan history record.
func (d *DB) GetHistoryByID(id int64) *ScanRecord {
	row := d.sql.QueryRow(
		`SELECT id, timestamp, tab, system, count, top_profit,
		 COALESCE(total_profit, 0), COALESCE(duration_ms, 0), COALESCE(params_json, '{}')
		 FROM scan_history WHERE id = ?`,
		id,
	)
	var r ScanRecord
	var paramsStr string
	if err := row.Scan(&r.ID, &r.Timestamp, &r.Tab, &r.System, &r.Count, &r.TopProfit, &r.TotalProfit, &r.DurationMs, &paramsStr); err != nil {
		return nil
	}
	r.Params = json.RawMessage(paramsStr)
	return &r
}

// DeleteHistory deletes a scan history record and its associated results.
func (d *DB) DeleteHistory(id int64) error {
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	tx.Exec("DELETE FROM flip_results WHERE scan_id = ?", id)
	tx.Exec("DELETE FROM regional_day_results WHERE scan_id = ?", id)
	tx.Exec("DELETE FROM contract_results WHERE scan_id = ?", id)
	tx.Exec("DELETE FROM station_results WHERE scan_id = ?", id)
	tx.Exec("DELETE FROM route_results WHERE scan_id = ?", id)
	tx.Exec("DELETE FROM scan_history WHERE id = ?", id)
	return tx.Commit()
}

// ClearHistory deletes all scan history records older than given days.
func (d *DB) ClearHistory(olderThanDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays).Format(time.RFC3339)

	// Get IDs to delete
	rows, err := d.sql.Query("SELECT id FROM scan_history WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) == 0 {
		return 0, nil
	}

	tx, err := d.sql.Begin()
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		tx.Exec("DELETE FROM flip_results WHERE scan_id = ?", id)
		tx.Exec("DELETE FROM regional_day_results WHERE scan_id = ?", id)
		tx.Exec("DELETE FROM contract_results WHERE scan_id = ?", id)
		tx.Exec("DELETE FROM station_results WHERE scan_id = ?", id)
		tx.Exec("DELETE FROM route_results WHERE scan_id = ?", id)
	}
	result, err := tx.Exec("DELETE FROM scan_history WHERE timestamp < ?", cutoff)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()
	count, _ := result.RowsAffected()
	return count, nil
}
