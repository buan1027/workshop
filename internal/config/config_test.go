package config

import "testing"

func TestLoadDefaultsToResetDatabaseOnStart(t *testing.T) {
	t.Setenv("RESET_DATABASE_ON_START", "")

	cfg := Load()

	if !cfg.ResetDatabaseOnStart {
		t.Fatal("expected reset database on start to be enabled by default")
	}
	if cfg.AuthMode != "admin-token" {
		t.Fatalf("expected admin-token auth mode by default, got %q", cfg.AuthMode)
	}
}

func TestLoadCanDisableResetDatabaseOnStart(t *testing.T) {
	t.Setenv("RESET_DATABASE_ON_START", "false")

	cfg := Load()

	if cfg.ResetDatabaseOnStart {
		t.Fatal("expected reset database on start to be disabled")
	}
}
