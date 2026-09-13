package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"neraca/internal/config"
	"neraca/internal/database"
)

func TestHandleHealth(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_health.db")

	db, err := database.Connect(dbPath)
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		Port:   "8088",
		DBPath: dbPath,
		Env:    "test",
	}

	router := NewRouter(cfg, db, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var res HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if res.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", res.Status)
	}
	if res.App != "neraca" {
		t.Errorf("expected app 'neraca', got '%s'", res.App)
	}
	if res.Database != "connected" {
		t.Errorf("expected database 'connected', got '%s'", res.Database)
	}
}
