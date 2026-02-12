package config

import (
	"os"
	"time"
)

type Config struct {
	Port                    string
	Env                     string
	DatabaseURL             string
	JWTSecret               string
	JWTAccessExpiry         time.Duration
	JWTRefreshExpiry        time.Duration
	GoogleMapsAPIKey        string
	FirebaseCredentialsPath string
	LocationStaleThreshold  time.Duration
	LocationLogRetentionDays int
	ReservationReminderMin  int
}

func Load() *Config {
	return &Config{
		Port:                    getEnv("PORT", "8080"),
		Env:                     getEnv("ENV", "development"),
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://driver:driver@localhost:5432/driver_db?sslmode=disable"),
		JWTSecret:               getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTAccessExpiry:         parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m")),
		JWTRefreshExpiry:        parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		GoogleMapsAPIKey:        getEnv("GOOGLE_MAPS_API_KEY", ""),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		LocationStaleThreshold:  parseDuration(getEnv("LOCATION_STALE_THRESHOLD", "2m")),
		LocationLogRetentionDays: 90,
		ReservationReminderMin:  30,
	}
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
