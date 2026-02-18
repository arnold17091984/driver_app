package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kento/driver/backend/internal/config"
	"github.com/kento/driver/backend/internal/db"
	"github.com/kento/driver/backend/internal/handler"
	"github.com/kento/driver/backend/internal/maps"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/internal/notify"
	"github.com/kento/driver/backend/internal/repository"
	"github.com/kento/driver/backend/internal/service"
)

func New(cfg *config.Config) (*http.Server, error) {
	// Database
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("database connection: %w", err)
	}

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}
	log.Println("database connected and migrations applied")

	// In production, check for seed users with well-known passwords
	if cfg.Env == "production" {
		var seedCount int
		err := database.Get(&seedCount,
			`SELECT COUNT(*) FROM users WHERE employee_id IN ('admin001','dispatch001','driver001','driver002','driver003','driver004','driver005','viewer001') AND is_active = true`)
		if err == nil && seedCount > 0 {
			return nil, fmt.Errorf("production: %d seed/test users are still active; deactivate or remove them before running in production", seedCount)
		}
	}

	// Repositories
	userRepo := repository.NewUserRepo(database)
	vehicleRepo := repository.NewVehicleRepo(database)
	dispatchRepo := repository.NewDispatchRepo(database)
	reservationRepo := repository.NewReservationRepo(database)
	conflictRepo := repository.NewConflictRepo(database)
	attendanceRepo := repository.NewAttendanceRepo(database)
	locationRepo := repository.NewLocationRepo(database)
	auditRepo := repository.NewAuditRepo(database)
	tokenRepo := repository.NewTokenRepo(database)

	// Notification service
	fcmSvc, err := notify.NewFCMService(cfg.FirebaseCredentialsPath, userRepo)
	if err != nil {
		log.Printf("[notify] FCM initialization failed: %v (continuing without push)", err)
		fcmSvc, _ = notify.NewFCMService("", userRepo)
	}

	// Services
	auditSvc := service.NewAuditService(auditRepo)
	tokenSvc := service.NewTokenService(tokenRepo)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	vehicleSvc := service.NewVehicleService(vehicleRepo, cfg.LocationStaleThreshold, auditSvc)
	attendanceSvc := service.NewAttendanceService(attendanceRepo, auditSvc)
	locationSvc := service.NewLocationService(locationRepo)
	dispatchSvc := service.NewDispatchService(dispatchRepo, vehicleRepo, auditSvc, cfg.LocationStaleThreshold, fcmSvc)
	reservationSvc := service.NewReservationService(reservationRepo, conflictRepo, auditSvc)
	conflictSvc := service.NewConflictService(conflictRepo, reservationRepo, auditSvc)
	bookingSvc := service.NewBookingService(dispatchSvc, reservationSvc, vehicleRepo, reservationRepo, auditSvc, fcmSvc)

	// Upload directory
	uploadDir := filepath.Join(".", "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	// Maps client
	mapsClient := maps.NewClient(cfg.GoogleMapsAPIKey)

	// Login rate limiter (5 failed attempts, 15 minute lockout)
	loginLimiter := middleware.NewLoginLimiter(5, 15*time.Minute)

	// Handlers
	authH := handler.NewAuthHandler(authSvc, tokenSvc, loginLimiter)
	vehicleH := handler.NewVehicleHandler(vehicleSvc, locationSvc, uploadDir)
	dispatchH := handler.NewDispatchHandler(dispatchSvc, vehicleSvc)
	reservationH := handler.NewReservationHandler(reservationSvc, authSvc)
	conflictH := handler.NewConflictHandler(conflictSvc, reservationSvc)
	attendanceH := handler.NewAttendanceHandler(attendanceSvc)
	locationH := handler.NewLocationHandler(locationSvc, vehicleSvc)
	adminH := handler.NewAdminHandler(userRepo, auditSvc)
	notifH := handler.NewNotificationHandler(userRepo)
	routeH := handler.NewRouteHandler(mapsClient)
	bookingH := handler.NewBookingHandler(bookingSvc, authSvc)
	passengerH := handler.NewPassengerHandler(authSvc, dispatchSvc, locationSvc, bookingSvc, loginLimiter)

	// Router
	router := buildRouter(
		cfg, tokenSvc,
		authH, vehicleH, dispatchH, reservationH, conflictH,
		attendanceH, locationH, adminH, notifH, routeH,
		bookingH, passengerH,
	)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// TLS support: use ListenAndServeTLS when certs are provided
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		srv.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return srv, nil
}
