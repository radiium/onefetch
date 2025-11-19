package handler

import (
	"dlbackend/internal/errors"
	"dlbackend/internal/model"
	"dlbackend/internal/service"
	"dlbackend/internal/utils"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// DownloadHandler handles HTTP requests for download operations.
type DownloadHandler interface {
	GetInfos(c *fiber.Ctx) error
	ListDownloads(c *fiber.Ctx) error
	CreateDownload(c *fiber.Ctx) error
	PauseDownload(c *fiber.Ctx) error
	ResumeDownload(c *fiber.Ctx) error
	CancelDownload(c *fiber.Ctx) error
	ArchiveDownload(c *fiber.Ctx) error
	DeleteDownload(c *fiber.Ctx) error
}

type downloadHandler struct {
	service service.DownloadService
}

// GetInfos get file info from 1fichier api
func (h *downloadHandler) GetInfos(c *fiber.Ctx) error {
	url, err := utils.ValidateNotEmpty("url", c.Query("url"))
	if err != nil {
		return errors.HandleError(c, err)
	}

	fileinfo, err := h.service.GetFileinfo(url)
	if err != nil {
		return errors.HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fileinfo)
}

// NewDownloadHandler creates a new DownloadHandler instance.
func NewDownloadHandler(service service.DownloadService) DownloadHandler {
	return &downloadHandler{service: service}
}

// ListDownloads get paginated downloads with filters
func (h *downloadHandler) ListDownloads(c *fiber.Ctx) error {
	status := c.Query("status", "")
	downloadType := c.Query("type", "")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var statusFilters []model.DownloadStatus
	if status != "" {
		statuses := strings.Split(status, ",")
		for _, s := range statuses {
			s = strings.TrimSpace(s)
			if s != "" {
				statusFilters = append(statusFilters, model.DownloadStatus(s))
			}
		}
	}

	var typeFilters []model.DownloadType
	if downloadType != "" {
		downloadTypes := strings.Split(downloadType, ",")
		for _, s := range downloadTypes {
			s = strings.TrimSpace(s)
			if s != "" {
				typeFilters = append(typeFilters, model.DownloadType(s))
			}
		}
	}

	downloads, total, err := h.service.ListDownloads(statusFilters, typeFilters, page, limit)
	if err != nil {
		return errors.HandleError(c,
			fmt.Errorf("failed to list downloads: status=%s; typeFilters=%s; page=%d; limit=%d; error=%s", statusFilters, typeFilters, page, limit, err.Error()),
		)
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	return c.JSON(fiber.Map{
		"data": downloads,
		"pagination": fiber.Map{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}

// CreateDownload create and start download
func (h *downloadHandler) CreateDownload(c *fiber.Ctx) error {
	// Validate request body
	var req model.CreateDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.HandleBodyParserError(c, err)
	}
	// Validate URL
	urlStr, err := utils.Validate1FichierURL(req.URL)
	if err != nil {
		return errors.HandleError(c, errors.BadRequest(err.Error()))
	}
	// Validate download type
	downloadType, err := utils.ValidateType(req.Type)
	if err != nil {
		return errors.HandleError(c, errors.BadRequest(err.Error()))
	}
	// Validate dirName
	fileDir, err := utils.ValidateDirName(*req.FileDir)
	if err != nil {
		return errors.HandleError(c, errors.BadRequest(err.Error()))
	}
	// Validate fileName
	fileName, err := utils.ValidateFileName(*req.FileName)
	if err != nil {
		return errors.HandleError(c, errors.BadRequest(err.Error()))
	}

	download, err := h.service.CreateDownload(urlStr, downloadType, fileDir, fileName)
	if err != nil {
		return errors.HandleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(download.Clone())
}

// PauseDownload pause a download
func (h *downloadHandler) PauseDownload(c *fiber.Ctx) error {
	// Validate id param
	id, err := utils.ValidateNotEmpty("id", c.Params("id"))
	if err != nil {
		return errors.HandleError(c, err)
	}

	if err := h.service.PauseDownload(id); err != nil {
		return errors.HandleError(c, fmt.Errorf("failed to pause download: %s %s", id, err.Error()))
	}

	return c.SendStatus(fiber.StatusOK)
}

// ResumeDownload resume a download
func (h *downloadHandler) ResumeDownload(c *fiber.Ctx) error {
	// Validate id param
	id, err := utils.ValidateNotEmpty("id", c.Params("id"))
	if err != nil {
		return errors.HandleError(c, err)
	}

	if err := h.service.ResumeDownload(id); err != nil {
		return errors.HandleError(c, fmt.Errorf("failed to resume download: %s %s", id, err.Error()))
	}

	return c.SendStatus(fiber.StatusOK)
}

// CancelDownload cancel a download
func (h *downloadHandler) CancelDownload(c *fiber.Ctx) error {
	// Validate id param
	id, err := utils.ValidateNotEmpty("id", c.Params("id"))
	if err != nil {
		return errors.HandleError(c, err)
	}

	if err := h.service.CancelDownload(id); err != nil {
		return errors.HandleError(c, fmt.Errorf("failed to cancel download: %s %s", id, err.Error()))
	}

	return c.SendStatus(fiber.StatusOK)
}

// ArchiveDownload archive a download
func (h *downloadHandler) ArchiveDownload(c *fiber.Ctx) error {
	// Validate id param
	id, err := utils.ValidateNotEmpty("id", c.Params("id"))
	if err != nil {
		return errors.HandleError(c, err)
	}

	if err := h.service.ArchiveDownload(id); err != nil {
		return errors.HandleError(c, fmt.Errorf("failed to archive download: %s %s", id, err.Error()))
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteDownload delete a download
func (h *downloadHandler) DeleteDownload(c *fiber.Ctx) error {
	// Validate id param
	id, err := utils.ValidateNotEmpty("id", c.Params("id"))
	if err != nil {
		return errors.HandleError(c, err)
	}

	if err := h.service.DeleteDownload(id); err != nil {
		return errors.HandleError(c, fmt.Errorf("failed to delete download: %s %s", id, err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}
