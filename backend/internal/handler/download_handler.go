package handler

import (
	"dlbackend/internal/model"
	"dlbackend/internal/service"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type DownloadHandler interface {
	CreateDownload(c *fiber.Ctx) error
	ListDownloads(c *fiber.Ctx) error
	PauseDownload(c *fiber.Ctx) error
	ResumeDownload(c *fiber.Ctx) error
	CancelDownload(c *fiber.Ctx) error
	ArchiveDownload(c *fiber.Ctx) error
	DeleteDownload(c *fiber.Ctx) error
}

type downloadHandler struct {
	service service.DownloadService
}

func NewDownloadHandler(service service.DownloadService) DownloadHandler {
	return &downloadHandler{service: service}
}

func (h *downloadHandler) CreateDownload(c *fiber.Ctx) error {
	var req model.CreateDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.URL == "" || req.Type == "" {
		return fiber.NewError(fiber.StatusBadRequest, "URL and type are required")
	}

	if err := h.validate1FichierURL(req.URL); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid 1fichier url")
	}

	downloadType := model.DownloadType(req.Type)
	if downloadType != model.TypeMovie && downloadType != model.TypeSerie {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid download type")
	}

	download, err := h.service.CreateDownload(req.URL, downloadType, *req.FileName, *req.FileDir)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(download)
}

func (h *downloadHandler) ListDownloads(c *fiber.Ctx) error {
	status := c.Query("status")
	downloadType := c.Query("type")
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
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list downloads")
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

func (h *downloadHandler) PauseDownload(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Download ID is required")
	}

	download, err := h.service.PauseDownload(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Download not found")
	}

	return c.Status(fiber.StatusOK).JSON(download)
}

func (h *downloadHandler) ResumeDownload(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Download ID is required")
	}

	download, err := h.service.ResumeDownload(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Download not found")
	}

	return c.Status(fiber.StatusOK).JSON(download)
}

func (h *downloadHandler) CancelDownload(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Download ID is required")
	}

	download, err := h.service.CancelDownload(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Download not found")
	}

	return c.Status(fiber.StatusOK).JSON(download)
}

func (h *downloadHandler) ArchiveDownload(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Download ID is required")
	}

	if err := h.service.ArchiveDownload(id); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Download not found")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *downloadHandler) DeleteDownload(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Download ID required")
	}

	if err := h.service.DeleteDownload(id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// generateUID génère un simple UID
func generateUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Valide l'URL 1fichier.com et retourne l'URL parsée
func (h *downloadHandler) validate1FichierURL(rawURL string) error {
	// Validation de base - l'URL ne doit pas être vide
	if strings.TrimSpace(rawURL) == "" {
		return errors.New("URL vide")
	}

	// Parse l'URL de manière sécurisée
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return errors.New("URL invalide")
	}

	// Vérifie que c'est bien un domaine 1fichier.com
	if !strings.HasSuffix(parsedURL.Host, "1fichier.com") {
		return errors.New("domaine non autorisé")
	}

	// Vérifie qu'il y a bien une query string
	if parsedURL.RawQuery == "" {
		return errors.New("aucun ID trouvé dans l'URL")
	}

	return nil
}
