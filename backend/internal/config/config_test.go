package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// Clear all env vars that Load reads
	envVars := []string{
		"PORT", "ENV", "DATABASE_URL", "JWT_SECRET",
		"JWT_ACCESS_EXPIRY", "JWT_REFRESH_EXPIRY",
		"GOOGLE_MAPS_API_KEY", "FIREBASE_CREDENTIALS_PATH",
		"LOCATION_STALE_THRESHOLD",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg := Load()

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
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "production")
	os.Setenv("DATABASE_URL", "postgres://prod:pass@db:5432/app")
	os.Setenv("JWT_SECRET", "super-secret")
	os.Setenv("JWT_ACCESS_EXPIRY", "30m")
	os.Setenv("JWT_REFRESH_EXPIRY", "720h")
	os.Setenv("LOCATION_STALE_THRESHOLD", "5m")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_ACCESS_EXPIRY")
		os.Unsetenv("JWT_REFRESH_EXPIRY")
		os.Unsetenv("LOCATION_STALE_THRESHOLD")
	}()

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.DatabaseURL != "postgres://prod:pass@db:5432/app" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://prod:pass@db:5432/app")
	}
	if cfg.JWTSecret != "super-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "super-secret")
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
}

func TestParseDurationInvalid(t *testing.T) {
	os.Setenv("JWT_ACCESS_EXPIRY", "not-a-duration")
	defer os.Unsetenv("JWT_ACCESS_EXPIRY")

	cfg := Load()

	// Invalid duration should fall back to 15 minutes
	if cfg.JWTAccessExpiry != 15*time.Minute {
		t.Errorf("JWTAccessExpiry = %v, want %v (fallback)", cfg.JWTAccessExpiry, 15*time.Minute)
	}
}
