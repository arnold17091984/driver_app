package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func clearEnv() {
	for _, v := range []string{
		"PORT", "ENV", "DATABASE_URL", "JWT_SECRET",
		"JWT_ACCESS_EXPIRY", "JWT_REFRESH_EXPIRY",
		"GOOGLE_MAPS_API_KEY", "FIREBASE_CREDENTIALS_PATH",
		"LOCATION_STALE_THRESHOLD", "CORS_ORIGINS",
		"RATE_LIMIT_RATE", "RATE_LIMIT_BURST",
	} {
		os.Unsetenv(v)
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
	if cfg.JWTAccessExpiry != 15*time.Minute {
		t.Errorf("JWTAccessExpiry = %v, want %v", cfg.JWTAccessExpiry, 15*time.Minute)
	}
	if cfg.JWTRefreshExpiry != 168*time.Hour {
		t.Errorf("JWTRefreshExpiry = %v, want %v", cfg.JWTRefreshExpiry, 168*time.Hour)
	}
	if cfg.LocationStaleThreshold != 2*time.Minute {
		t.Errorf("LocationStaleThreshold = %v, want %v", cfg.LocationStaleThreshold, 2*time.Minute)
	}
	if cfg.LocationLogRetentionDays != 90 {
		t.Errorf("LocationLogRetentionDays = %d, want %d", cfg.LocationLogRetentionDays, 90)
	}
	if cfg.ReservationReminderMin != 30 {
		t.Errorf("ReservationReminderMin = %d, want %d", cfg.ReservationReminderMin, 30)
	}
	if len(cfg.CORSOrigins) != 1 || cfg.CORSOrigins[0] != "*" {
		t.Errorf("CORSOrigins = %v, want [*]", cfg.CORSOrigins)
	}
	if cfg.RateLimitRate != 20 {
		t.Errorf("RateLimitRate = %v, want 20", cfg.RateLimitRate)
	}
	if cfg.RateLimitBurst != 40 {
		t.Errorf("RateLimitBurst = %v, want 40", cfg.RateLimitBurst)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "production")
	os.Setenv("DATABASE_URL", "postgres://prod:pass@db:5432/app")
	os.Setenv("JWT_SECRET", "a-very-long-production-secret-key-12345678")
	os.Setenv("JWT_ACCESS_EXPIRY", "30m")
	os.Setenv("JWT_REFRESH_EXPIRY", "720h")
	os.Setenv("LOCATION_STALE_THRESHOLD", "5m")
	os.Setenv("CORS_ORIGINS", "https://app.example.com,https://admin.example.com")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.DatabaseURL != "postgres://prod:pass@db:5432/app" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://prod:pass@db:5432/app")
	}
	if cfg.JWTSecret != "a-very-long-production-secret-key-12345678" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "a-very-long-production-secret-key-12345678")
	}
	if cfg.JWTAccessExpiry != 30*time.Minute {
		t.Errorf("JWTAccessExpiry = %v, want %v", cfg.JWTAccessExpiry, 30*time.Minute)
	}
	if cfg.JWTRefreshExpiry != 720*time.Hour {
		t.Errorf("JWTRefreshExpiry = %v, want %v", cfg.JWTRefreshExpiry, 720*time.Hour)
	}
	if cfg.LocationStaleThreshold != 5*time.Minute {
		t.Errorf("LocationStaleThreshold = %v, want %v", cfg.LocationStaleThreshold, 5*time.Minute)
	}
	if len(cfg.CORSOrigins) != 2 || cfg.CORSOrigins[0] != "https://app.example.com" {
		t.Errorf("CORSOrigins = %v, want [https://app.example.com, https://admin.example.com]", cfg.CORSOrigins)
	}
}

func TestParseDurationInvalid(t *testing.T) {
	clearEnv()
	os.Setenv("JWT_ACCESS_EXPIRY", "not-a-duration")
	defer os.Unsetenv("JWT_ACCESS_EXPIRY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.JWTAccessExpiry != 15*time.Minute {
		t.Errorf("JWTAccessExpiry = %v, want %v (fallback)", cfg.JWTAccessExpiry, 15*time.Minute)
	}
}

func TestProductionShortJWTSecret(t *testing.T) {
	clearEnv()
	os.Setenv("ENV", "production")
	os.Setenv("JWT_SECRET", "short")
	os.Setenv("DATABASE_URL", "postgres://prod:pass@db:5432/app")
	os.Setenv("CORS_ORIGINS", "https://example.com")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for short JWT secret in production")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET must be at least 32 characters") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestProductionWeakJWTSecret(t *testing.T) {
	clearEnv()
	os.Setenv("ENV", "production")
	os.Setenv("JWT_SECRET", "change-me-in-production-padding-extra-chars")
	os.Setenv("DATABASE_URL", "postgres://prod:pass@db:5432/app")
	os.Setenv("CORS_ORIGINS", "https://example.com")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for known weak JWT secret in production")
	}
}

func TestProductionLocalhostDB(t *testing.T) {
	clearEnv()
	os.Setenv("ENV", "production")
	os.Setenv("JWT_SECRET", "a-very-long-production-secret-key-12345678")
	os.Setenv("DATABASE_URL", "postgres://driver:driver@localhost:5432/driver_db")
	os.Setenv("CORS_ORIGINS", "https://example.com")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for localhost DATABASE_URL in production")
	}
	if !strings.Contains(err.Error(), "localhost") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestProductionWildcardCORS(t *testing.T) {
	clearEnv()
	os.Setenv("ENV", "production")
	os.Setenv("JWT_SECRET", "a-very-long-production-secret-key-12345678")
	os.Setenv("DATABASE_URL", "postgres://prod:pass@db:5432/app")
	os.Setenv("CORS_ORIGINS", "*")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for wildcard CORS in production")
	}
	if !strings.Contains(err.Error(), "CORS_ORIGINS") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDevelopmentAllowsWeakConfig(t *testing.T) {
	clearEnv()
	// Default ENV=development should pass even with weak defaults
	cfg, err := Load()
	if err != nil {
		t.Fatalf("development should not fail validation: %v", err)
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want development", cfg.Env)
	}
}

func TestParseCORSOrigins(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", []string{"*"}},
		{"*", []string{"*"}},
		{"https://a.com", []string{"https://a.com"}},
		{"https://a.com, https://b.com", []string{"https://a.com", "https://b.com"}},
	}
	for _, tt := range tests {
		got := parseCORSOrigins(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("parseCORSOrigins(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseCORSOrigins(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}
