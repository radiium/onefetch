package utils

import (
	"dlbackend/internal/handler"
	"dlbackend/internal/repository"
	"dlbackend/internal/service"
	"dlbackend/pkg/config"
	"dlbackend/pkg/database"
	"dlbackend/pkg/sse"
)

type ServiceContainer struct {
	Config          *config.Config
	DB              *database.Database
	SSEManager      sse.Manager
	DownloadHandler handler.DownloadHandler
	SettingsHandler handler.SettingsHandler
	FileinfoHandler handler.FileinfoHandler
}

// NewServiceContainer initialise le container avec toutes les dépendances
func NewServiceContainer(cfg *config.Config, db *database.Database, sseManager sse.Manager) *ServiceContainer {
	// Repositories
	downloadRepo := repository.NewDownloadRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Services
	downloadService := service.NewDownloadService(downloadRepo, settingsRepo, sseManager)
	settingsService := service.NewSettingsService(settingsRepo)
	fileinfoService := service.NewFileinfoService(settingsRepo)

	// Handlers
	downloadHandler := handler.NewDownloadHandler(downloadService)
	settingsHandler := handler.NewSettingsHandler(settingsService)
	fileinfoHandler := handler.NewFileinfoHandler(fileinfoService)

	return &ServiceContainer{
		Config:          cfg,
		DB:              db,
		SSEManager:      sseManager,
		DownloadHandler: downloadHandler,
		SettingsHandler: settingsHandler,
		FileinfoHandler: fileinfoHandler,
	}
}
