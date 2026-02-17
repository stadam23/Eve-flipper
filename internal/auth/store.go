package auth

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// Session represents a stored auth session.
type Session struct {
	CharacterID   int64
	CharacterName string
	AccessToken   string
	RefreshToken  string
	ExpiresAt     time.Time
	Active        bool
}

// SessionStore handles session persistence in SQLite.
type SessionStore struct {
	db *sql.DB
}

const defaultUserID = "default"

// NewSessionStore creates a store backed by the given SQL database.
func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

func normalizeUserID(userID string) string {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return defaultUserID
	}
	return trimmed
}

// Save stores or updates a character session while preserving active selection.
// If there is no active character yet, this session becomes active.
func (s *SessionStore) Save(sess *Session) error {
	return s.SaveForUser(defaultUserID, sess)
}

// SaveForUser stores or updates a character session for the given user while preserving active selection.
// If there is no active character yet, this session becomes active.
func (s *SessionStore) SaveForUser(userID string, sess *Session) error {
	userID = normalizeUserID(userID)
	if sess == nil {
		return fmt.Errorf("nil session")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO auth_session (user_id, character_id, character_name, access_token, refresh_token, expires_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, 0)
		ON CONFLICT(user_id, character_id) DO UPDATE SET
			character_name = excluded.character_name,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			expires_at = excluded.expires_at`,
		userID, sess.CharacterID, sess.CharacterName, sess.AccessToken, sess.RefreshToken, sess.ExpiresAt.Unix(),
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE auth_session
		   SET is_active = 1
		 WHERE user_id = ? AND character_id = ?
		   AND NOT EXISTS (SELECT 1 FROM auth_session WHERE user_id = ? AND is_active = 1)`,
		userID, sess.CharacterID, userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SaveAndActivate stores or updates a character session and marks it as active.
func (s *SessionStore) SaveAndActivate(sess *Session) error {
	return s.SaveAndActivateForUser(defaultUserID, sess)
}

// SaveAndActivateForUser stores or updates a character session for the given user and marks it as active.
func (s *SessionStore) SaveAndActivateForUser(userID string, sess *Session) error {
	userID = normalizeUserID(userID)
	if sess == nil {
		return fmt.Errorf("nil session")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE auth_session SET is_active = 0 WHERE user_id = ?`, userID); err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO auth_session (user_id, character_id, character_name, access_token, refresh_token, expires_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, 1)
		ON CONFLICT(user_id, character_id) DO UPDATE SET
			character_name = excluded.character_name,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			expires_at = excluded.expires_at,
			is_active = 1`,
		userID, sess.CharacterID, sess.CharacterName, sess.AccessToken, sess.RefreshToken, sess.ExpiresAt.Unix(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Get returns the active session, or nil if none.
func (s *SessionStore) Get() *Session {
	return s.GetForUser(defaultUserID)
}

// GetForUser returns the active session for the given user, or nil if none.
func (s *SessionStore) GetForUser(userID string) *Session {
	userID = normalizeUserID(userID)

	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM auth_session WHERE user_id = ?`, userID).Scan(&count); err != nil || count == 0 {
		return nil
	}

	if sess := s.querySession(`
		SELECT character_id, character_name, access_token, refresh_token, expires_at, is_active
		FROM auth_session
		WHERE user_id = ? AND is_active = 1
		LIMIT 1`, userID); sess != nil {
		return sess
	}
	// Fallback for legacy/edge states: return first session even if no active flag.
	return s.querySession(`
		SELECT character_id, character_name, access_token, refresh_token, expires_at, is_active
		FROM auth_session
		WHERE user_id = ?
		ORDER BY character_name ASC, character_id ASC
		LIMIT 1`, userID)
}

// GetByCharacterID returns a specific character session.
func (s *SessionStore) GetByCharacterID(characterID int64) *Session {
	return s.GetByCharacterIDForUser(defaultUserID, characterID)
}

// GetByCharacterIDForUser returns a specific character session for the given user.
func (s *SessionStore) GetByCharacterIDForUser(userID string, characterID int64) *Session {
	userID = normalizeUserID(userID)

	return s.querySession(`
		SELECT character_id, character_name, access_token, refresh_token, expires_at, is_active
		FROM auth_session
		WHERE user_id = ? AND character_id = ?
		LIMIT 1`, userID, characterID)
}

// List returns all stored character sessions (active first).
func (s *SessionStore) List() []*Session {
	return s.ListForUser(defaultUserID)
}

// ListForUser returns all stored character sessions for the given user (active first).
func (s *SessionStore) ListForUser(userID string) []*Session {
	userID = normalizeUserID(userID)

	rows, err := s.db.Query(`
		SELECT character_id, character_name, access_token, refresh_token, expires_at, is_active
		FROM auth_session
		WHERE user_id = ?
		ORDER BY is_active DESC, character_name ASC, character_id ASC`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []*Session
	for rows.Next() {
		var sess Session
		var expiresUnix int64
		var activeInt int
		if err := rows.Scan(&sess.CharacterID, &sess.CharacterName, &sess.AccessToken, &sess.RefreshToken, &expiresUnix, &activeInt); err != nil {
			continue
		}
		sess.ExpiresAt = time.Unix(expiresUnix, 0)
		sess.Active = activeInt == 1
		out = append(out, &sess)
	}
	return out
}

func (s *SessionStore) querySession(query string, args ...interface{}) *Session {
	var sess Session
	var expiresUnix int64
	var activeInt int
	err := s.db.QueryRow(query, args...).
		Scan(&sess.CharacterID, &sess.CharacterName, &sess.AccessToken, &sess.RefreshToken, &expiresUnix, &activeInt)
	if err != nil {
		return nil
	}
	sess.ExpiresAt = time.Unix(expiresUnix, 0)
	sess.Active = activeInt == 1
	return &sess
}

// SetActive marks a stored character as active.
func (s *SessionStore) SetActive(characterID int64) error {
	return s.SetActiveForUser(defaultUserID, characterID)
}

// SetActiveForUser marks a stored character as active for the given user.
func (s *SessionStore) SetActiveForUser(userID string, characterID int64) error {
	userID = normalizeUserID(userID)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE auth_session SET is_active = 0 WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}
	_, _ = res.RowsAffected()

	res, err = tx.Exec(`UPDATE auth_session SET is_active = 1 WHERE user_id = ? AND character_id = ?`, userID, characterID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("character not found")
	}

	return tx.Commit()
}

// Delete removes all stored sessions.
func (s *SessionStore) Delete() {
	s.DeleteForUser(defaultUserID)
}

// DeleteForUser removes all stored sessions for the given user.
func (s *SessionStore) DeleteForUser(userID string) {
	userID = normalizeUserID(userID)
	s.db.Exec("DELETE FROM auth_session WHERE user_id = ?", userID)
}

// DeleteByCharacterID removes a specific character session.
func (s *SessionStore) DeleteByCharacterID(characterID int64) error {
	return s.DeleteByCharacterIDForUser(defaultUserID, characterID)
}

// DeleteByCharacterIDForUser removes a specific character session for the given user.
func (s *SessionStore) DeleteByCharacterIDForUser(userID string, characterID int64) error {
	userID = normalizeUserID(userID)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var wasActive int
	err = tx.QueryRow(`SELECT is_active FROM auth_session WHERE user_id = ? AND character_id = ?`, userID, characterID).Scan(&wasActive)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM auth_session WHERE user_id = ? AND character_id = ?`, userID, characterID); err != nil {
		return err
	}

	if wasActive == 1 {
		if _, err := tx.Exec(`
			UPDATE auth_session
			   SET is_active = 1
			 WHERE user_id = ? AND character_id = (
				SELECT character_id
				  FROM auth_session
				 WHERE user_id = ?
				 ORDER BY character_name ASC, character_id ASC
				 LIMIT 1
			 )`, userID, userID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// EnsureValidToken returns a valid access token, refreshing if needed.
func (s *SessionStore) EnsureValidToken(sso *SSOConfig) (string, error) {
	return s.EnsureValidTokenForUser(sso, defaultUserID)
}

// EnsureValidTokenForUser returns a valid access token for the active character of the given user.
func (s *SessionStore) EnsureValidTokenForUser(sso *SSOConfig, userID string) (string, error) {
	userID = normalizeUserID(userID)

	sess := s.GetForUser(userID)
	if sess == nil {
		return "", fmt.Errorf("not logged in")
	}
	return s.ensureValidTokenForSession(userID, sess, sso)
}

// EnsureValidTokenForCharacter returns a valid access token for the given character.
func (s *SessionStore) EnsureValidTokenForCharacter(sso *SSOConfig, characterID int64) (string, error) {
	return s.EnsureValidTokenForUserCharacter(sso, defaultUserID, characterID)
}

// EnsureValidTokenForUserCharacter returns a valid access token for the given user and character.
func (s *SessionStore) EnsureValidTokenForUserCharacter(sso *SSOConfig, userID string, characterID int64) (string, error) {
	userID = normalizeUserID(userID)

	sess := s.GetByCharacterIDForUser(userID, characterID)
	if sess == nil {
		return "", fmt.Errorf("character not logged in")
	}
	return s.ensureValidTokenForSession(userID, sess, sso)
}

func (s *SessionStore) ensureValidTokenForSession(userID string, sess *Session, sso *SSOConfig) (string, error) {
	if sess == nil {
		return "", fmt.Errorf("not logged in")
	}

	// If token is still valid (with 60s buffer), return it
	if time.Now().Before(sess.ExpiresAt.Add(-60 * time.Second)) {
		return sess.AccessToken, nil
	}
	if sso == nil {
		return "", fmt.Errorf("sso not configured")
	}

	// Refresh the token
	log.Printf("[AUTH] Refreshing token for %s", sess.CharacterName)
	tok, err := sso.RefreshToken(sess.RefreshToken)
	if err != nil {
		_ = s.DeleteByCharacterIDForUser(userID, sess.CharacterID)
		return "", fmt.Errorf("refresh failed: %w", err)
	}

	sess.AccessToken = tok.AccessToken
	sess.RefreshToken = tok.RefreshToken
	sess.ExpiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	if err := s.SaveForUser(userID, sess); err != nil {
		return "", fmt.Errorf("save session: %w", err)
	}

	return sess.AccessToken, nil
}
