package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kento/driver/backend/internal/config"
	"github.com/kento/driver/backend/internal/server"
)

func main() {
	// Load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("server starting on %s (env: %s)", srv.Addr, cfg.Env)

	var listenErr error
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		log.Println("TLS enabled")
		listenErr = srv.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
	} else {
		if cfg.Env == "production" {
			log.Println("WARNING: running without TLS in production")
		}
		listenErr = srv.ListenAndServe()
	}
	if listenErr != nil && listenErr != http.ErrServerClosed {
		log.Printf("server stopped: %v", listenErr)
	}
}
