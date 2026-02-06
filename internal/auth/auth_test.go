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
			id              INTEGER PRIMARY KEY DEFAULT 1,
			character_id    INTEGER NOT NULL,
			character_name  TEXT NOT NULL,
			access_token    TEXT NOT NULL,
			refresh_token   TEXT NOT NULL,
			expires_at      INTEGER NOT NULL
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
	if err := store.Save(sess); err != nil {
		t.Fatalf("Save: %v", err)
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

	store.Delete()
	if store.Get() != nil {
		t.Error("Get() after Delete should return nil")
	}
}
