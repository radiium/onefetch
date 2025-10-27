package handler

import (
	"dlbackend/internal/model"
	"dlbackend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type SettingsHandler interface {
	GetSettings(c *fiber.Ctx) error
	UpdateSettings(c *fiber.Ctx) error
}

type settingsHandler struct {
	service service.SettingsService
}

func NewSettingsHandler(service service.SettingsService) SettingsHandler {
	return &settingsHandler{service: service}
}

func (h *settingsHandler) GetSettings(c *fiber.Ctx) error {
	settings, err := h.service.GetSettings()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve settings")
	}

	return c.JSON(settings)
}

func (h *settingsHandler) UpdateSettings(c *fiber.Ctx) error {
	var settings model.UpdateSettingsRequest
	if err := c.BodyParser(&settings); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if settings.APIKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "APIKey are required")
	}

	updated, err := h.service.UpdateSettings(&settings)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update settings")
	}

	return c.JSON(updated)
}
