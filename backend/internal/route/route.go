package route

import (
	"dlbackend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, container *utils.ServiceContainer) {
	api := app.Group("/api")

	// Settings routes
	settings := api.Group("/settings")
	settings.Get("/", container.SettingsHandler.GetSettings)
	settings.Patch("/", container.SettingsHandler.UpdateSettings)

	// Fileinfo routes
	files := api.Group("/fileinfo")
	files.Get("", container.FileinfoHandler.Get)

	// Download routes
	downloads := api.Group("/downloads")
	downloads.Post("/", container.DownloadHandler.CreateDownload)
	downloads.Get("/", container.DownloadHandler.ListDownloads)
	downloads.Post("/:id/pause", container.DownloadHandler.PauseDownload)
	downloads.Post("/:id/resume", container.DownloadHandler.ResumeDownload)
	downloads.Post("/:id/cancel", container.DownloadHandler.CancelDownload)
	downloads.Post("/:id/archive", container.DownloadHandler.ArchiveDownload)
	downloads.Delete("/:id", container.DownloadHandler.DeleteDownload)

	// Download SSE routes
	downloads.Get("/streams", container.SSEManager.Handler)

	// Static webapp
	if container.Config.IsProd() {
		app.Static("/", "./web")
		app.Get("/new", func(c *fiber.Ctx) error {
			return c.SendFile("./web/new.html")
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
