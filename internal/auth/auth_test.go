package auth

import (
	"database/sql"
	"encoding/base64"
	"net/url"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestBuildAuthURL_Exact(t *testing.T) {
	c := &SSOConfig{
		ClientID:    "test-client",
		CallbackURL: "http://localhost:13370/callback",
		Scopes:      "esi-markets.read_character_orders.v1",
	}
	u := c.BuildAuthURL("abc123")
	if !strings.HasPrefix(u, "https://login.eveonline.com/v2/oauth/authorize?") {
		t.Errorf("BuildAuthURL prefix wrong: %q", u)
	}
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	q := parsed.Query()
	if q.Get("response_type") != "code" {
		t.Errorf("response_type = %q", q.Get("response_type"))
	}
	if q.Get("client_id") != "test-client" {
		t.Errorf("client_id = %q", q.Get("client_id"))
	}
	if q.Get("redirect_uri") != "http://localhost:13370/callback" {
		t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
	}
	if q.Get("scope") != c.Scopes {
		t.Errorf("scope = %q", q.Get("scope"))
	}
	if q.Get("state") != "abc123" {
		t.Errorf("state = %q", q.Get("state"))
	}
}

func TestGenerateState_LengthAndEncoding(t *testing.T) {
	s := GenerateState()
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		t.Errorf("GenerateState not valid base64 URL: %v", err)
	}
	if len(decoded) != 16 {
		t.Errorf("GenerateState decoded length = %d, want 16", len(decoded))
	}
	// Two calls should differ (with very high probability)
	s2 := GenerateState()
	if s == s2 {
		t.Error("GenerateState should return different values")
	}
}

func TestSessionStore_SaveGetDelete(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	_, err = sqlDB.Exec(`
		CREATE TABLE auth_session (
			user_id         TEXT NOT NULL,
			character_id    INTEGER NOT NULL,
			character_name  TEXT NOT NULL,
			access_token    TEXT NOT NULL,
			refresh_token   TEXT NOT NULL,
			expires_at      INTEGER NOT NULL,
			is_active       INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (user_id, character_id)
		)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	store := NewSessionStore(sqlDB)

	if store.Get() != nil {
		t.Error("Get() on empty store should return nil")
	}

	sess := &Session{
		CharacterID:   12345,
		CharacterName: "Test Char",
		AccessToken:   "at",
		RefreshToken:  "rt",
		ExpiresAt:     time.Now().Add(time.Hour),
	}
	if err := store.SaveAndActivate(sess); err != nil {
		t.Fatalf("SaveAndActivate: %v", err)
	}

	got := store.Get()
	if got == nil {
		t.Fatal("Get() after Save returned nil")
	}
	if got.CharacterID != 12345 || got.CharacterName != "Test Char" {
		t.Errorf("Get() = %+v", got)
	}
	if got.AccessToken != "at" || got.RefreshToken != "rt" {
		t.Errorf("tokens = %q / %q", got.AccessToken, got.RefreshToken)
	}
	if !got.Active {
		t.Errorf("expected active session")
	}

	second := &Session{
		CharacterID:   67890,
		CharacterName: "Alt Char",
		AccessToken:   "at2",
		RefreshToken:  "rt2",
		ExpiresAt:     time.Now().Add(time.Hour),
	}
	if err := store.Save(second); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	list := store.List()
	if len(list) != 2 {
		t.Fatalf("List() len = %d, want 2", len(list))
	}
	if list[0].CharacterID != 12345 || !list[0].Active {
		t.Fatalf("expected first list entry to be active character 12345, got %+v", list[0])
	}

	if err := store.SetActive(67890); err != nil {
		t.Fatalf("SetActive: %v", err)
	}
	if active := store.Get(); active == nil || active.CharacterID != 67890 {
		t.Fatalf("active after SetActive = %+v, want character 67890", active)
	}

	if err := store.DeleteByCharacterID(67890); err != nil {
		t.Fatalf("DeleteByCharacterID: %v", err)
	}
	if active := store.Get(); active == nil || active.CharacterID != 12345 {
		t.Fatalf("active after deleting 67890 = %+v, want character 12345", active)
	}

	store.Delete()
	if store.Get() != nil {
		t.Error("Get() after Delete should return nil")
	}
}

func TestSessionStore_UserIsolation(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	_, err = sqlDB.Exec(`
		CREATE TABLE auth_session (
			user_id         TEXT NOT NULL,
			character_id    INTEGER NOT NULL,
			character_name  TEXT NOT NULL,
			access_token    TEXT NOT NULL,
			refresh_token   TEXT NOT NULL,
			expires_at      INTEGER NOT NULL,
			is_active       INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (user_id, character_id)
		)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	store := NewSessionStore(sqlDB)

	if err := store.SaveAndActivateForUser("u1", &Session{
		CharacterID:   1001,
		CharacterName: "User One",
		AccessToken:   "at1",
		RefreshToken:  "rt1",
		ExpiresAt:     time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("SaveAndActivateForUser u1: %v", err)
	}
	if err := store.SaveAndActivateForUser("u2", &Session{
		CharacterID:   2002,
		CharacterName: "User Two",
		AccessToken:   "at2",
		RefreshToken:  "rt2",
		ExpiresAt:     time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("SaveAndActivateForUser u2: %v", err)
	}

	if got := store.GetForUser("u1"); got == nil || got.CharacterID != 1001 {
		t.Fatalf("GetForUser(u1) = %+v, want character 1001", got)
	}
	if got := store.GetForUser("u2"); got == nil || got.CharacterID != 2002 {
		t.Fatalf("GetForUser(u2) = %+v, want character 2002", got)
	}
	if got := store.Get(); got != nil {
		t.Fatalf("default Get() should be empty for non-default users, got %+v", got)
	}
}

func newSessionStoreForTokenTest(t *testing.T) *SessionStore {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	_, err = sqlDB.Exec(`
		CREATE TABLE auth_session (
			user_id         TEXT NOT NULL,
			character_id    INTEGER NOT NULL,
			character_name  TEXT NOT NULL,
			access_token    TEXT NOT NULL,
			refresh_token   TEXT NOT NULL,
			expires_at      INTEGER NOT NULL,
			is_active       INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (user_id, character_id)
		)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	return NewSessionStore(sqlDB)
}

func TestSessionStore_EnsureValidTokenForUser_UsesUnexpiredTokenWithoutSSO(t *testing.T) {
	store := newSessionStoreForTokenTest(t)
	err := store.SaveAndActivateForUser("u1", &Session{
		CharacterID:   101,
		CharacterName: "Pilot One",
		AccessToken:   "access-token",
		RefreshToken:  "refresh-token",
		ExpiresAt:     time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("SaveAndActivateForUser: %v", err)
	}

	token, err := store.EnsureValidTokenForUser(nil, "u1")
	if err != nil {
		t.Fatalf("EnsureValidTokenForUser: %v", err)
	}
	if token != "access-token" {
		t.Fatalf("token = %q, want access-token", token)
	}
}

func TestSessionStore_EnsureValidTokenForUser_ExpiredTokenRequiresSSO(t *testing.T) {
	store := newSessionStoreForTokenTest(t)
	err := store.SaveAndActivateForUser("u1", &Session{
		CharacterID:   101,
		CharacterName: "Pilot One",
		AccessToken:   "access-token",
		RefreshToken:  "refresh-token",
		ExpiresAt:     time.Now().Add(-2 * time.Minute),
	})
	if err != nil {
		t.Fatalf("SaveAndActivateForUser: %v", err)
	}

	_, err = store.EnsureValidTokenForUser(nil, "u1")
	if err == nil {
		t.Fatal("expected error for expired token without sso")
	}
	if !strings.Contains(err.Error(), "sso not configured") {
		t.Fatalf("error = %v, want contains %q", err, "sso not configured")
	}
}
