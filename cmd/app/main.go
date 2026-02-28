package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	apphttp "go-prod-app/internal/http"
	"go-prod-app/internal/logger"
	"go-prod-app/internal/metrics"
	"go-prod-app/internal/repository"
	"go-prod-app/internal/service"
)

func main() {

	// =========================
	// Logger
	// =========================
	log := logger.New()

	// =========================
	// Load ENV
	// =========================
	_ = godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Error("missing DB_DSN")
		os.Exit(1)
	}

	// =========================
	// Connect DB
	// =========================
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error("failed to open db", "error", err)
		os.Exit(1)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Error("failed to ping db", "error", err)
		os.Exit(1)
	}

	log.Info("database connected")

	// =========================
	// Init Metrics
	// =========================
	metrics.Init()

	// =========================
	// Wire Dependencies
	// =========================
	userRepo := repository.NewPostgresUserRepository(db)
	userService := service.NewUserService(userRepo, userRepo)

	// =========================
	// Start HTTP Server
	// =========================
	server := apphttp.StartServer(userService, log)

	// =========================
	// Graceful Shutdown
	// =========================
	waitForShutdown(log)

	log.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown failed", "error", err)
	}

	if err := db.Close(); err != nil {
		log.Error("error closing db", "error", err)
	}

	log.Info("shutdown complete")
}

func waitForShutdown(log *slog.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info("received shutdown signal", "signal", sig.String())
}
