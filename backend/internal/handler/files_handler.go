package handler

import (
	"dlbackend/internal/errors"
	"dlbackend/internal/model"
	"dlbackend/internal/service"
	"dlbackend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// FilesHandler handles HTTP requests for filesystem operations.
type FilesHandler interface {
	Get(c *fiber.Ctx) error
	Post(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type filesHandler struct {
	service service.FilesService
}

// NewFilesHandler creates a new FilesHandler instance.
func NewFilesHandler(service service.FilesService) FilesHandler {
	return &filesHandler{service: service}
}

// Get get filesystem tree representation of download directory
func (h *filesHandler) Get(c *fiber.Ctx) error {
	tree, err := h.service.GetDir()
	if err != nil {
		return errors.HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(tree)
}

// Post create new folder in filesystem tree of download directory
func (h *filesHandler) Post(c *fiber.Ctx) error {
	// Validate request body
	var body model.CreateDirRequest
	if err := c.BodyParser(&body); err != nil {
		return errors.HandleBodyParserError(c, err)
	}
	// Validate path
	path, err := utils.ValidatePath(body.Path)
	if err != nil {
		return errors.HandleError(c, errors.BadRequest(err.Error()))
	}
	// Validate dirName
	dirName, err := utils.ValidateDirName(body.DirName)
	if err != nil {
		return errors.HandleError(c, errors.BadRequest(err.Error()))
	}

	tree, err := h.service.CreateDir(path, dirName)
	if err != nil {
		return errors.HandleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(tree)
}

// Delete delete file or folder in filesystem tree of download directory
func (h *filesHandler) Delete(c *fiber.Ctx) error {
	path, err := utils.ValidatePath(c.Query("path"))
	if err != nil {
		return errors.HandleError(c, errors.BadRequest(err.Error()))
	}

	tree, err := h.service.DeleteDir(path)
	if err != nil {
		return errors.HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(tree)
}
