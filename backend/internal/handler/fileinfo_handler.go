package handler

import (
	"dlbackend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type FileinfoHandler interface {
	Get(c *fiber.Ctx) error
}

type fileinfoHandler struct {
	service service.FileinfoService
}

func NewFileinfoHandler(service service.FileinfoService) FileinfoHandler {
	return &fileinfoHandler{service: service}
}

func (h *fileinfoHandler) Get(c *fiber.Ctx) error {
	url := c.Query("url", "")
	if url == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid URL")
	}

	fileinfo, err := h.service.GetFileinfo(url)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fileinfo)
}
