package db

import (
	"fmt"
	"strings"
	"time"
)

const (
	TradeStateModeDone    = "done"
	TradeStateModeIgnored = "ignored"
)

type TradeState struct {
	UserID        string `json:"user_id"`
	Tab           string `json:"tab"`
	TypeID        int32  `json:"type_id"`
	StationID     int64  `json:"station_id"`
	RegionID      int32  `json:"region_id"`
	Mode          string `json:"mode"`
	UntilRevision int64  `json:"until_revision"`
	UpdatedAt     string `json:"updated_at"`
}

type TradeStateKey struct {
	TypeID    int32 `json:"type_id"`
	StationID int64 `json:"station_id"`
	RegionID  int32 `json:"region_id"`
}

func normalizeTradeStateTab(tab string) string {
	tab = strings.ToLower(strings.TrimSpace(tab))
	if tab == "" {
		return "station"
	}
	return tab
}

func normalizeTradeStateMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case TradeStateModeDone:
		return TradeStateModeDone
	case TradeStateModeIgnored:
		return TradeStateModeIgnored
	default:
		return ""
	}
}

func (d *DB) UpsertTradeStateForUser(userID string, state TradeState) error {
	userID = normalizeUserID(userID)
	state.Tab = normalizeTradeStateTab(state.Tab)
	state.Mode = normalizeTradeStateMode(state.Mode)
	if state.Mode == "" {
		return fmt.Errorf("invalid trade-state mode")
	}
	if state.TypeID <= 0 || state.StationID <= 0 {
		return fmt.Errorf("type_id and station_id must be positive")
	}

	updatedAt := time.Now().UTC().Format(time.RFC3339)
	_, err := d.sql.Exec(`
		INSERT INTO user_trade_state (
			user_id, tab, type_id, station_id, region_id, mode, until_revision, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, tab, type_id, station_id, region_id)
		DO UPDATE SET
			mode = excluded.mode,
			until_revision = excluded.until_revision,
			updated_at = excluded.updated_at
	`,
		userID, state.Tab, state.TypeID, state.StationID, state.RegionID, state.Mode, state.UntilRevision, updatedAt,
	)
	return err
}

func (d *DB) ListTradeStatesForUser(userID, tab string) ([]TradeState, error) {
	userID = normalizeUserID(userID)
	tab = normalizeTradeStateTab(tab)

	rows, err := d.sql.Query(`
		SELECT user_id, tab, type_id, station_id, region_id, mode, until_revision, updated_at
		  FROM user_trade_state
		 WHERE user_id = ? AND tab = ?
		 ORDER BY updated_at DESC
	`, userID, tab)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TradeState
	for rows.Next() {
		var s TradeState
		if err := rows.Scan(
			&s.UserID, &s.Tab, &s.TypeID, &s.StationID, &s.RegionID, &s.Mode, &s.UntilRevision, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *DB) DeleteTradeStatesForUser(userID, tab string, keys []TradeStateKey) error {
	if len(keys) == 0 {
		return nil
	}
	userID = normalizeUserID(userID)
	tab = normalizeTradeStateTab(tab)

	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		DELETE FROM user_trade_state
		 WHERE user_id = ? AND tab = ? AND type_id = ? AND station_id = ? AND region_id = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, key := range keys {
		if key.TypeID <= 0 || key.StationID <= 0 {
			continue
		}
		if _, err := stmt.Exec(userID, tab, key.TypeID, key.StationID, key.RegionID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *DB) ClearTradeStatesForUser(userID, tab, mode string) (int64, error) {
	userID = normalizeUserID(userID)
	tab = normalizeTradeStateTab(tab)
	mode = normalizeTradeStateMode(mode)

	if mode == "" {
		res, err := d.sql.Exec(`
			DELETE FROM user_trade_state WHERE user_id = ? AND tab = ?
		`, userID, tab)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}

	res, err := d.sql.Exec(`
		DELETE FROM user_trade_state WHERE user_id = ? AND tab = ? AND mode = ?
	`, userID, tab, mode)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *DB) DeleteExpiredDoneTradeStatesForUser(userID, tab string, currentRevision int64) (int64, error) {
	if currentRevision <= 0 {
		return 0, nil
	}
	userID = normalizeUserID(userID)
	tab = normalizeTradeStateTab(tab)

	res, err := d.sql.Exec(`
		DELETE FROM user_trade_state
		 WHERE user_id = ?
		   AND tab = ?
		   AND mode = ?
		   AND until_revision > 0
		   AND until_revision < ?
	`, userID, tab, TradeStateModeDone, currentRevision)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
