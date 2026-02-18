package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// knownWeakSecrets is a blocklist of default/weak JWT secrets that must not
// be used in production.
var knownWeakSecrets = map[string]bool{
	"dev-secret-change-me":   true,
	"change-me-in-production": true,
	"secret":                  true,
	"supersecret":             true,
	"password":                true,
}

type Config struct {
	Port                     string
	Env                      string
	DatabaseURL              string
	JWTSecret                string
	JWTAccessExpiry          time.Duration
	JWTRefreshExpiry         time.Duration
	GoogleMapsAPIKey         string
	FirebaseCredentialsPath  string
	LocationStaleThreshold   time.Duration
	LocationLogRetentionDays int
	ReservationReminderMin   int
	CORSOrigins              []string
	RateLimitRate            float64
	RateLimitBurst           int
	TLSCert                  string
	TLSKey                   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                     getEnv("PORT", "8080"),
		Env:                      getEnv("ENV", "development"),
		DatabaseURL:              getEnv("DATABASE_URL", "postgres://driver:driver@localhost:5432/driver_db?sslmode=disable"),
		JWTSecret:                getEnv("JWT_SECRET", ""),
		JWTAccessExpiry:          parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m")),
		JWTRefreshExpiry:         parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		GoogleMapsAPIKey:         getEnv("GOOGLE_MAPS_API_KEY", ""),
		FirebaseCredentialsPath:  getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		LocationStaleThreshold:   parseDuration(getEnv("LOCATION_STALE_THRESHOLD", "2m")),
		LocationLogRetentionDays: 90,
		ReservationReminderMin:   30,
		CORSOrigins:              parseCORSOrigins(getEnv("CORS_ORIGINS", "http://localhost:5173")),
		RateLimitRate:            parseFloat(getEnv("RATE_LIMIT_RATE", "20")),
		RateLimitBurst:           parseInt(getEnv("RATE_LIMIT_BURST", "40")),
		TLSCert:                  getEnv("TLS_CERT", ""),
		TLSKey:                   getEnv("TLS_KEY", ""),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	// JWT_SECRET is always required (no default fallback)
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}

	// Block known weak secrets in all environments
	if knownWeakSecrets[c.JWTSecret] {
		if c.Env == "production" {
			return fmt.Errorf("production: JWT_SECRET is a known weak value; set a strong secret")
		}
	}

	if c.Env != "production" {
		return nil
	}

	// Production: JWT secret must be >= 32 chars and not a known weak value
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("production: JWT_SECRET must be at least 32 characters (got %d)", len(c.JWTSecret))
	}
	for weak := range knownWeakSecrets {
		if strings.HasPrefix(c.JWTSecret, weak) {
			return fmt.Errorf("production: JWT_SECRET starts with a known default value; set a strong secret")
		}
	}

	// Production: do not allow localhost database
	lower := strings.ToLower(c.DatabaseURL)
	if strings.Contains(lower, "localhost") || strings.Contains(lower, "127.0.0.1") {
		return fmt.Errorf("production: DATABASE_URL must not point to localhost")
	}

	// Production: require SSL for database connection
	if strings.Contains(lower, "sslmode=disable") {
		return fmt.Errorf("production: DATABASE_URL must not use sslmode=disable")
	}

	// Production: CORS wildcard not allowed
	for _, o := range c.CORSOrigins {
		if o == "*" {
			return fmt.Errorf("production: CORS_ORIGINS must not be wildcard (*); specify explicit origins")
		}
	}

	return nil
}

func parseCORSOrigins(raw string) []string {
	if raw == "" {
		return []string{"*"}
	}
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 20
	}
	return f
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 40
	}
	return i
}
