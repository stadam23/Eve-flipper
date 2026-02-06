package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// DemandRegion represents cached demand data for a region.
type DemandRegion struct {
	RegionID      int32     `json:"region_id"`
	RegionName    string    `json:"region_name"`
	HotScore      float64   `json:"hot_score"`
	Status        string    `json:"status"`
	KillsToday    int64     `json:"kills_today"`
	KillsBaseline int64     `json:"kills_baseline"`
	ISKDestroyed  float64   `json:"isk_destroyed"`
	ActivePlayers int       `json:"active_players"`
	TopShips      []string  `json:"top_ships"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DemandItem represents cached demand data for an item.
type DemandItem struct {
	RegionID     int32     `json:"region_id"`
	TypeID       int32     `json:"type_id"`
	TypeName     string    `json:"type_name"`
	GroupID      int32     `json:"group_id"`
	GroupName    string    `json:"group_name"`
	LossesPerDay int64     `json:"losses_per_day"`
	DemandScore  float64   `json:"demand_score"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SaveDemandRegion saves or updates demand data for a region.
func (d *DB) SaveDemandRegion(region *DemandRegion) error {
	topShipsJSON, _ := json.Marshal(region.TopShips)

	_, err := d.sql.Exec(`
		INSERT OR REPLACE INTO demand_region_cache 
		(region_id, region_name, hot_score, status, kills_today, kills_baseline, 
		 isk_destroyed, active_players, top_ships, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, region.RegionID, region.RegionName, region.HotScore, region.Status,
		region.KillsToday, region.KillsBaseline, region.ISKDestroyed,
		region.ActivePlayers, string(topShipsJSON), time.Now().Format(time.RFC3339))

	return err
}

// GetDemandRegions returns cached demand data for all regions.
func (d *DB) GetDemandRegions() ([]DemandRegion, error) {
	rows, err := d.sql.Query(`
		SELECT region_id, region_name, hot_score, status, kills_today, kills_baseline,
		       isk_destroyed, active_players, top_ships, updated_at
		FROM demand_region_cache
		ORDER BY hot_score DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []DemandRegion
	for rows.Next() {
		var r DemandRegion
		var topShipsJSON string
		var updatedAtStr string

		err := rows.Scan(&r.RegionID, &r.RegionName, &r.HotScore, &r.Status,
			&r.KillsToday, &r.KillsBaseline, &r.ISKDestroyed, &r.ActivePlayers,
			&topShipsJSON, &updatedAtStr)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(topShipsJSON), &r.TopShips)
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		regions = append(regions, r)
	}

	return regions, nil
}

// GetDemandRegion returns cached demand data for a specific region.
func (d *DB) GetDemandRegion(regionID int32) (*DemandRegion, error) {
	var r DemandRegion
	var topShipsJSON string
	var updatedAtStr string

	err := d.sql.QueryRow(`
		SELECT region_id, region_name, hot_score, status, kills_today, kills_baseline,
		       isk_destroyed, active_players, top_ships, updated_at
		FROM demand_region_cache
		WHERE region_id = ?
	`, regionID).Scan(&r.RegionID, &r.RegionName, &r.HotScore, &r.Status,
		&r.KillsToday, &r.KillsBaseline, &r.ISKDestroyed, &r.ActivePlayers,
		&topShipsJSON, &updatedAtStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(topShipsJSON), &r.TopShips)
	r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &r, nil
}

// GetHotZones returns regions with elevated activity (hot_score > 1.2).
func (d *DB) GetHotZones(limit int) ([]DemandRegion, error) {
	query := `
		SELECT region_id, region_name, hot_score, status, kills_today, kills_baseline,
		       isk_destroyed, active_players, top_ships, updated_at
		FROM demand_region_cache
		WHERE hot_score >= 1.2
		ORDER BY hot_score DESC
	`
	if limit > 0 {
		query += " LIMIT ?"
	}

	var rows *sql.Rows
	var err error
	if limit > 0 {
		rows, err = d.sql.Query(query, limit)
	} else {
		rows, err = d.sql.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []DemandRegion
	for rows.Next() {
		var r DemandRegion
		var topShipsJSON string
		var updatedAtStr string

		err := rows.Scan(&r.RegionID, &r.RegionName, &r.HotScore, &r.Status,
			&r.KillsToday, &r.KillsBaseline, &r.ISKDestroyed, &r.ActivePlayers,
			&topShipsJSON, &updatedAtStr)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(topShipsJSON), &r.TopShips)
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		regions = append(regions, r)
	}

	return regions, nil
}

// SaveDemandItem saves or updates demand data for an item.
func (d *DB) SaveDemandItem(item *DemandItem) error {
	_, err := d.sql.Exec(`
		INSERT OR REPLACE INTO demand_item_cache 
		(region_id, type_id, type_name, group_id, group_name, losses_per_day, demand_score, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, item.RegionID, item.TypeID, item.TypeName, item.GroupID, item.GroupName,
		item.LossesPerDay, item.DemandScore, time.Now().Format(time.RFC3339))

	return err
}

// GetTopDemandItems returns items with highest demand scores.
func (d *DB) GetTopDemandItems(regionID int32, limit int) ([]DemandItem, error) {
	query := `
		SELECT region_id, type_id, type_name, group_id, group_name, losses_per_day, demand_score, updated_at
		FROM demand_item_cache
	`
	args := []interface{}{}

	if regionID > 0 {
		query += " WHERE region_id = ?"
		args = append(args, regionID)
	}

	query += " ORDER BY demand_score DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := d.sql.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DemandItem
	for rows.Next() {
		var item DemandItem
		var updatedAtStr string

		err := rows.Scan(&item.RegionID, &item.TypeID, &item.TypeName, &item.GroupID,
			&item.GroupName, &item.LossesPerDay, &item.DemandScore, &updatedAtStr)
		if err != nil {
			continue
		}

		item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		items = append(items, item)
	}

	return items, nil
}

// IsDemandCacheFresh checks if the demand cache is recent enough.
func (d *DB) IsDemandCacheFresh(maxAge time.Duration) bool {
	var updatedAtStr string
	err := d.sql.QueryRow(`
		SELECT updated_at FROM demand_region_cache ORDER BY updated_at DESC LIMIT 1
	`).Scan(&updatedAtStr)

	if err != nil {
		return false
	}

	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return false
	}

	return time.Since(updatedAt) < maxAge
}

// ClearDemandCache clears all cached demand data.
func (d *DB) ClearDemandCache() error {
	_, err := d.sql.Exec(`
		DELETE FROM demand_region_cache;
		DELETE FROM demand_item_cache;
	`)
	return err
}

// FittingDemandItem represents a cached fitting demand item.
type FittingDemandItem struct {
	RegionID       int32     `json:"region_id"`
	TypeID         int32     `json:"type_id"`
	TypeName       string    `json:"type_name"`
	Category       string    `json:"category"`
	TotalDestroyed int64     `json:"total_destroyed"`
	KillmailCount  int       `json:"killmail_count"`
	AvgPerKillmail float64   `json:"avg_per_killmail"`
	EstDailyDemand float64   `json:"est_daily_demand"`
	SampledKills   int       `json:"sampled_kills"`
	TotalKills24h  int       `json:"total_kills_24h"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SaveFittingDemandProfile saves fitting demand data for a region.
func (d *DB) SaveFittingDemandProfile(regionID int32, items []FittingDemandItem) error {
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear old data for this region
	if _, err := tx.Exec(`DELETE FROM demand_fitting_cache WHERE region_id = ?`, regionID); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO demand_fitting_cache
		(region_id, type_id, type_name, category, total_destroyed, killmail_count,
		 avg_per_killmail, est_daily_demand, sampled_kills, total_kills_24h, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Format(time.RFC3339)
	for _, item := range items {
		if _, err := stmt.Exec(
			regionID, item.TypeID, item.TypeName, item.Category,
			item.TotalDestroyed, item.KillmailCount, item.AvgPerKillmail,
			item.EstDailyDemand, item.SampledKills, item.TotalKills24h, now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetFittingDemandProfile returns cached fitting demand data for a region.
func (d *DB) GetFittingDemandProfile(regionID int32) ([]FittingDemandItem, error) {
	rows, err := d.sql.Query(`
		SELECT region_id, type_id, type_name, category, total_destroyed, killmail_count,
		       avg_per_killmail, est_daily_demand, sampled_kills, total_kills_24h, updated_at
		FROM demand_fitting_cache
		WHERE region_id = ?
		ORDER BY est_daily_demand DESC
	`, regionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []FittingDemandItem
	for rows.Next() {
		var item FittingDemandItem
		var updatedAtStr string
		err := rows.Scan(&item.RegionID, &item.TypeID, &item.TypeName, &item.Category,
			&item.TotalDestroyed, &item.KillmailCount, &item.AvgPerKillmail,
			&item.EstDailyDemand, &item.SampledKills, &item.TotalKills24h, &updatedAtStr)
		if err != nil {
			continue
		}
		item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		items = append(items, item)
	}
	return items, nil
}

// IsFittingProfileFresh checks if fitting demand data for a region is recent enough.
func (d *DB) IsFittingProfileFresh(regionID int32, maxAge time.Duration) bool {
	var updatedAtStr string
	err := d.sql.QueryRow(`
		SELECT updated_at FROM demand_fitting_cache WHERE region_id = ? LIMIT 1
	`, regionID).Scan(&updatedAtStr)
	if err != nil {
		return false
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return false
	}
	return time.Since(updatedAt) < maxAge
}
