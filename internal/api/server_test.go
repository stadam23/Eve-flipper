package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eve-flipper/internal/config"
	"eve-flipper/internal/esi"
)

// GET /api/status is not tested here because it calls esi.Client.HealthCheck() which performs a real HTTP request.

func TestHandleGetConfig_ReturnsConfig(t *testing.T) {
	cfg := &config.Config{SystemName: "Jita", CargoCapacity: 10000}
	srv := NewServer(cfg, &esi.Client{}, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/config status = %d, want 200", rec.Code)
	}
	var out config.Config
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if out.SystemName != "Jita" || out.CargoCapacity != 10000 {
		t.Errorf("config = %+v", out)
	}
}
