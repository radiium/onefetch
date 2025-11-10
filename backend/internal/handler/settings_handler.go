package handler

import (
	"dlbackend/internal/errors"
	"dlbackend/internal/model"
	"dlbackend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// FilesHandler handles HTTP requests for settings operations.
type SettingsHandler interface {
	GetSettings(c *fiber.Ctx) error
	UpdateSettings(c *fiber.Ctx) error
}

type settingsHandler struct {
	service service.SettingsService
}

// NewSettingsHandler creates a new SettingsHandler instance.
func NewSettingsHandler(service service.SettingsService) SettingsHandler {
	return &settingsHandler{service: service}
}

// GetSettings get current Settings
func (h *settingsHandler) GetSettings(c *fiber.Ctx) error {
	settings, err := h.service.GetSettings()
	if err != nil {
		return errors.HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(settings)
}

// UpdateSettings update Settings
func (h *settingsHandler) UpdateSettings(c *fiber.Ctx) error {
	// Validate request body
	var settings model.UpdateSettingsRequest
	if err := c.BodyParser(&settings); err != nil {
		return errors.HandleBodyParserError(c, err)
	}

	updated, err := h.service.UpdateSettings(&settings)
	if err != nil {
		return errors.HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(updated)
}
