package config

import (
	"testing"
	"time"
)

var allEnvKeys = []string{
	"LINKUP_ADDR",
	"LINKUP_READ_HEADER_TIMEOUT",
	"LINKUP_READ_TIMEOUT",
	"LINKUP_WRITE_TIMEOUT",
	"LINKUP_IDLE_TIMEOUT",
	"LINKUP_SHUTDOWN_TIMEOUT",
	"LINKUP_COMMAND_TIMEOUT",
	"PROFILE_API_ADDR",
	"PROFILE_API_READ_HEADER_TIMEOUT",
	"PROFILE_API_READ_TIMEOUT",
	"PROFILE_API_WRITE_TIMEOUT",
	"PROFILE_API_IDLE_TIMEOUT",
	"PROFILE_API_SHUTDOWN_TIMEOUT",
	"PROFILE_BASE_URL",
	"PROFILE_TIMEOUT",
	"PROFILE_MAX_BODY",
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range allEnvKeys {
		t.Setenv(key, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.LinkUp.Address != ":8080" {
		t.Errorf("LinkUp.Address=%q", cfg.LinkUp.Address)
	}
	if cfg.LinkUp.ReadHeaderTimeout != 2*time.Second ||
		cfg.LinkUp.ReadTimeout != 5*time.Second ||
		cfg.LinkUp.WriteTimeout != 5*time.Second ||
		cfg.LinkUp.IdleTimeout != 30*time.Second ||
		cfg.LinkUp.ShutdownTimeout != 10*time.Second {
		t.Errorf("LinkUp server budgets: %+v", cfg.LinkUp)
	}
	if cfg.CommandTimeout != 800*time.Millisecond {
		t.Errorf("CommandTimeout=%s", cfg.CommandTimeout)
	}
	if cfg.ProfileAPI.Address != "127.0.0.1:8081" {
		t.Errorf("ProfileAPI.Address=%q", cfg.ProfileAPI.Address)
	}
	if cfg.ProfileBaseURL != "http://127.0.0.1:8081" {
		t.Errorf("ProfileBaseURL=%q", cfg.ProfileBaseURL)
	}
	if cfg.ProfileTimeout != 500*time.Millisecond {
		t.Errorf("ProfileTimeout=%s", cfg.ProfileTimeout)
	}
	if cfg.ProfileMaxBody != 64<<10 {
		t.Errorf("ProfileMaxBody=%d", cfg.ProfileMaxBody)
	}
}

func TestLoadFromEnv(t *testing.T) {
	clearEnv(t)

	t.Setenv("LINKUP_ADDR", "127.0.0.1:9999")
	t.Setenv("LINKUP_READ_HEADER_TIMEOUT", "3s")
	t.Setenv("LINKUP_READ_TIMEOUT", "6s")
	t.Setenv("LINKUP_WRITE_TIMEOUT", "7s")
	t.Setenv("LINKUP_IDLE_TIMEOUT", "1m")
	t.Setenv("LINKUP_SHUTDOWN_TIMEOUT", "2500ms")
	t.Setenv("LINKUP_COMMAND_TIMEOUT", "1500ms")
	t.Setenv("PROFILE_API_ADDR", "127.0.0.1:8082")
	t.Setenv("PROFILE_BASE_URL", "https://profiles.example.com")
	t.Setenv("PROFILE_TIMEOUT", "2s")
	t.Setenv("PROFILE_MAX_BODY", "1024")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.LinkUp.Address != "127.0.0.1:9999" ||
		cfg.LinkUp.ReadHeaderTimeout != 3*time.Second ||
		cfg.LinkUp.ReadTimeout != 6*time.Second ||
		cfg.LinkUp.WriteTimeout != 7*time.Second ||
		cfg.LinkUp.IdleTimeout != time.Minute ||
		cfg.LinkUp.ShutdownTimeout != 2500*time.Millisecond {
		t.Errorf("LinkUp: %+v", cfg.LinkUp)
	}
	if cfg.CommandTimeout != 1500*time.Millisecond {
		t.Errorf("CommandTimeout=%s", cfg.CommandTimeout)
	}
	if cfg.ProfileAPI.Address != "127.0.0.1:8082" {
		t.Errorf("ProfileAPI.Address=%q", cfg.ProfileAPI.Address)
	}
	if cfg.ProfileBaseURL != "https://profiles.example.com" ||
		cfg.ProfileTimeout != 2*time.Second ||
		cfg.ProfileMaxBody != 1024 {
		t.Errorf("Profile client: %+v", cfg)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	for _, tt := range []struct {
		name  string
		key   string
		value string
	}{
		{"bad duration", "LINKUP_READ_TIMEOUT", "soon"},
		{"zero duration", "LINKUP_WRITE_TIMEOUT", "0s"},
		{"negative duration", "LINKUP_IDLE_TIMEOUT", "-5s"},
		{"zero command budget", "LINKUP_COMMAND_TIMEOUT", "0"},
		{"bad shutdown budget", "LINKUP_SHUTDOWN_TIMEOUT", "never"},
		{"empty address", "LINKUP_ADDR", "   "},
		{"bad profile timeout", "PROFILE_TIMEOUT", "abc"},
		{"zero profile timeout", "PROFILE_TIMEOUT", "0ms"},
		{"bad profile body limit", "PROFILE_MAX_BODY", "lots"},
		{"negative profile body limit", "PROFILE_MAX_BODY", "-64"},
		{"bad base URL scheme", "PROFILE_BASE_URL", "ftp://127.0.0.1:8081"},
		{"base URL without host", "PROFILE_BASE_URL", "http://"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv(tt.key, tt.value)

			if _, err := Load(); err == nil {
				t.Fatalf("Load with %s=%q: want error", tt.key, tt.value)
			}
		})
	}
}
