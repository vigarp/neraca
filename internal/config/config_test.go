package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Pastikan env bersih saat pengujian default
	os.Unsetenv("PORT")
	os.Unsetenv("DB_PATH")
	os.Unsetenv("ENV")

	cfg := Load()

	if cfg.Port != "8088" {
		t.Errorf("expected default Port '8088', got '%s'", cfg.Port)
	}
	if cfg.DBPath != "./data/neraca.db" {
		t.Errorf("expected default DBPath './data/neraca.db', got '%s'", cfg.DBPath)
	}
	if cfg.Env != "development" {
		t.Errorf("expected default Env 'development', got '%s'", cfg.Env)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("DB_PATH", "/tmp/test.db")
	t.Setenv("ENV", "production")

	cfg := Load()

	if cfg.Port != "9999" {
		t.Errorf("expected custom Port '9999', got '%s'", cfg.Port)
	}
	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("expected custom DBPath '/tmp/test.db', got '%s'", cfg.DBPath)
	}
	if cfg.Env != "production" {
		t.Errorf("expected custom Env 'production', got '%s'", cfg.Env)
	}
}
