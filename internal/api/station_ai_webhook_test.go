package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateGitHubWebhookSignature(t *testing.T) {
	body := []byte(`{"ok":true}`)
	secret := "top-secret"
	signature := signGitHubWebhookBody(body, secret)

	if !validateGitHubWebhookSignature(body, signature, secret) {
		t.Fatal("expected valid signature to pass verification")
	}
	if validateGitHubWebhookSignature(body, signature, "wrong-secret") {
		t.Fatal("expected invalid secret to fail verification")
	}
	if validateGitHubWebhookSignature(body, "sha256=zz", secret) {
		t.Fatal("expected malformed signature hex to fail verification")
	}
}

func TestHandleInternalWikiGollumWebhook_RejectsInvalidSignature(t *testing.T) {
	t.Setenv(stationAIWikiWebhookSecretEnv, "secret-1")
	srv := &Server{}

	body := `{"repository":{"full_name":"ilyaux/Eve-flipper"},"pages":[{"page_name":"Station-Trading"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/internal/wiki/gollum", strings.NewReader(body))
	req.Header.Set("X-GitHub-Event", "gollum")
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandleInternalWikiGollumWebhook_AcceptsPingAndGollum(t *testing.T) {
	t.Setenv(stationAIWikiWebhookSecretEnv, "")
	srv := &Server{}

	pingReq := httptest.NewRequest(http.MethodPost, "/api/internal/wiki/gollum", strings.NewReader(`{}`))
	pingReq.Header.Set("X-GitHub-Event", "ping")
	pingRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(pingRec, pingReq)
	if pingRec.Code != http.StatusAccepted {
		t.Fatalf("ping status = %d, want %d", pingRec.Code, http.StatusAccepted)
	}

	gollumBody := `{"repository":{"full_name":"ilyaux/Eve-flipper"},"pages":[{"page_name":"Station-Trading","action":"edited"}]}`
	gollumReq := httptest.NewRequest(http.MethodPost, "/api/internal/wiki/gollum", strings.NewReader(gollumBody))
	gollumReq.Header.Set("X-GitHub-Event", "gollum")
	gollumRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(gollumRec, gollumReq)
	if gollumRec.Code != http.StatusAccepted {
		t.Fatalf("gollum status = %d, want %d", gollumRec.Code, http.StatusAccepted)
	}
}

func TestHandleInternalWikiGollumWebhook_AcceptsValidSignature(t *testing.T) {
	secret := "secret-2"
	t.Setenv(stationAIWikiWebhookSecretEnv, secret)
	srv := &Server{}

	body := []byte(`{"repository":{"full_name":"ilyaux/Eve-flipper"},"pages":[{"page_name":"Home","action":"edited"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/internal/wiki/gollum", strings.NewReader(string(body)))
	req.Header.Set("X-GitHub-Event", "gollum")
	req.Header.Set("X-Hub-Signature-256", signGitHubWebhookBody(body, secret))

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
}

func signGitHubWebhookBody(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
