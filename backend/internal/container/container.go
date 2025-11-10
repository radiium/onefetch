package container

import (
	"dlbackend/internal/database"
	"dlbackend/internal/handler"
	"dlbackend/internal/repository"
	"dlbackend/internal/service"
	"dlbackend/pkg/sse"
)

// Container holds all application dependencies for dependency injection.
type Container struct {
	DB              *database.Database
	SSEManager      sse.Manager
	DownloadHandler handler.DownloadHandler
	SettingsHandler handler.SettingsHandler
	FilesHandler    handler.FilesHandler
}

// New creates a Container with all dependencies wired up.
func New(db *database.Database, sseManager sse.Manager) *Container {
	// Repositories
	downloadRepo := repository.NewDownloadRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Services
	filesService := service.NewFilesService()
	downloadService := service.NewDownloadService(downloadRepo, settingsRepo, filesService, sseManager)
	settingsService := service.NewSettingsService(settingsRepo)

	// Handlers
	downloadHandler := handler.NewDownloadHandler(downloadService)
	settingsHandler := handler.NewSettingsHandler(settingsService)
	filesHandler := handler.NewFilesHandler(filesService)

	return &Container{
		DB:              db,
		SSEManager:      sseManager,
		DownloadHandler: downloadHandler,
		SettingsHandler: settingsHandler,
		FilesHandler:    filesHandler,
	}
}
