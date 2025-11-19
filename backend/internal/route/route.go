package route

import (
	"dlbackend/internal/config"
	"dlbackend/internal/container"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, container *container.Container) {
	api := app.Group("/api")

	// Settings routes
	settings := api.Group("/settings")
	settings.Get("/", container.SettingsHandler.GetSettings)
	settings.Patch("/", container.SettingsHandler.UpdateSettings)

	// Download routes
	downloads := api.Group("/downloads")
	downloads.Get("/infos", container.DownloadHandler.GetInfos)
	downloads.Get("/", container.DownloadHandler.ListDownloads)
	downloads.Post("/", container.DownloadHandler.CreateDownload)
	downloads.Post("/:id/pause", container.DownloadHandler.PauseDownload)
	downloads.Post("/:id/resume", container.DownloadHandler.ResumeDownload)
	downloads.Post("/:id/cancel", container.DownloadHandler.CancelDownload)
	downloads.Post("/:id/archive", container.DownloadHandler.ArchiveDownload)
	downloads.Delete("/:id", container.DownloadHandler.DeleteDownload)

	// Download SSE routes
	downloads.Get("/streams", container.SSEManager.Handler)

	// Files
	files := api.Group("/files")
	files.Get("/", container.FilesHandler.Get)
	files.Post("/", container.FilesHandler.Post)
	files.Delete("/", container.FilesHandler.Delete)

	// Static webapp
	if config.Cfg.IsProd() {
		app.Static("/", "./web")
		app.Get("/active", func(c *fiber.Ctx) error {
			return c.SendFile("./web/active.html")
		})
		app.Get("/files", func(c *fiber.Ctx) error {
			return c.SendFile("./web/files.html")
		})
		app.Get("/settings", func(c *fiber.Ctx) error {
			return c.SendFile("./web/settings.html")
		})
		app.Get("/history", func(c *fiber.Ctx) error {
			return c.SendFile("./web/history.html")
		})
		app.Get("/.well-known/appspecific/com.chrome.devtools.json", func(c *fiber.Ctx) error {
			return c.SendString("Go away, Chrome DevTools!")
		})
	}
}
