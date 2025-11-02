package main

import (
	"context"
	"dlbackend/internal/model"
	"dlbackend/internal/route"
	"dlbackend/internal/utils"
	"dlbackend/pkg/config"
	"dlbackend/pkg/database"
	"dlbackend/pkg/sse"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func initDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatalf("Failed to create directory: %s %v", path, err)
		}
	}
}

func main() {
	// Load configuration
	cfg := config.New()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "DLBackend",
	})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New())

	// Initialize data directory
	initDir(cfg.DataPath)
	initDir(cfg.DLPath)
	initDir(filepath.Join(cfg.DLPath, model.TypeMovie.Dir()))
	initDir(filepath.Join(cfg.DLPath, model.TypeSerie.Dir()))

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Failed to close database: %v", err)
		}
	}()

	// Initialize SSE
	sseManager := sse.New(sse.ManagerConfig{
		Name: "Download",
	})
	defer func() {
		if err := sseManager.Close(); err != nil {
			log.Fatalf("Failed to close sse: %v", err)
		}
	}()

	container := utils.NewServiceContainer(cfg, db, sseManager)
	// Initialize routes
	route.SetupRoutes(app, container)

	// Start server in goroutine
	go func() {
		log.Infof("🚀 Starting server on port %s", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for server shutdown signal
	waitForShutdown(app)
}

func waitForShutdown(app *fiber.App) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Warnf("Error during shutdown: %v", err)
	}

	log.Info("Server shutdown complete.")
}
