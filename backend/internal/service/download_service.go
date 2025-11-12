package service

import (
	"dlbackend/internal/config"
	"dlbackend/internal/errors"
	"dlbackend/internal/model"
	"dlbackend/internal/repository"
	"dlbackend/internal/utils"
	"dlbackend/pkg/client"
	"dlbackend/pkg/sse"
	"dlbackend/pkg/worker"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type DownloadService interface {
	GetFileinfo(fileURL string) (*model.DownloadInfoResponse, error)
	ListDownloads(status []model.DownloadStatus, downloadType []model.DownloadType, page, limit int) ([]model.Download, int64, error)
	CreateDownload(fileURL string, downloadType model.DownloadType, dirName string, fileName string) (*model.Download, error)
	PauseDownload(id string) (*model.Download, error)
	ResumeDownload(id string) (*model.Download, error)
	CancelDownload(id string) (*model.Download, error)
	ArchiveDownload(id string) error
	DeleteDownload(id string) error
}

type downloadService struct {
	downloadRepo  repository.DownloadRepository
	settingsRepo  repository.SettingsRepository
	filesService  FilesService
	sseManager    sse.Manager
	workerManager *worker.Manager
}

func NewDownloadService(
	downloadRepo repository.DownloadRepository,
	settingsRepo repository.SettingsRepository,
	filesService FilesService,
	sseManager sse.Manager,
) DownloadService {
	return &downloadService{
		downloadRepo:  downloadRepo,
		settingsRepo:  settingsRepo,
		filesService:  filesService,
		sseManager:    sseManager,
		workerManager: worker.NewManager(),
	}
}

func (ds *downloadService) GetFileinfo(fileURL string) (*model.DownloadInfoResponse, error) {
	settings, err := ds.settingsRepo.Get()
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to load settings: %v", err))
	}
	if settings.APIKey1fichier == "" {
		return nil, errors.Internal("1fichier API key not configured")
	}

	oneFichierClient := client.NewOneFichierClient(settings.APIKey1fichier)
	fileinfo, err := oneFichierClient.GetFileInfo(fileURL)
	if err != nil {
		log.Error(err)
		return nil, errors.Internal("failed to retrieve file info from 1fichier API")
	}

	moviePath := filepath.Join(config.Cfg.DLPath, model.TypeMovie.Dir())
	movieDirectories, err := utils.BuildDirTreeAsList(moviePath)
	if err != nil {
		return nil, fmt.Errorf("Get movies directories error: %w", err)
	}

	seriePath := filepath.Join(config.Cfg.DLPath, model.TypeSerie.Dir())
	serieDirectories, err := utils.BuildDirTreeAsList(seriePath)
	if err != nil {
		return nil, fmt.Errorf("Get series directories error: %w", err)
	}

	return &model.DownloadInfoResponse{
		Fileinfo: *fileinfo,
		Directories: map[model.DownloadType][]string{
			model.TypeMovie: movieDirectories,
			model.TypeSerie: serieDirectories,
		},
	}, nil
}

func (ds *downloadService) ListDownloads(status []model.DownloadStatus, downloadTypes []model.DownloadType, page, limit int) ([]model.Download, int64, error) {
	return ds.downloadRepo.List(status, downloadTypes, page, limit)
}

func (ds *downloadService) CreateDownload(fileURL string, downloadType model.DownloadType, customDirName string, customFileName string) (*model.Download, error) {
	settings, err := ds.settingsRepo.Get()
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to load settings: %v", err))
	}
	if settings.APIKey1fichier == "" {
		return nil, errors.Internal("1fichier API key not configured")
	}

	// Extraire l'ID du fichier
	fileID, err := ds.extract1fichierFileID(fileURL)
	if err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	downloadPath := filepath.Join(config.Cfg.DLPath, downloadType.Dir(), customDirName)

	download := &model.Download{
		ID:             uuid.New().String(),
		FileURL:        fileURL,
		FileID:         fileID,
		CustomFileName: &customFileName,
		Type:           downloadType,
		Status:         model.StatusPending,
		Progress:       0,
		DownloadPath:   downloadPath,
		RetryCount:     0,
		IsArchived:     false,
	}

	if err := ds.downloadRepo.Create(download); err != nil {
		return nil, err
	}

	// Démarrer le téléchargement dans une goroutine
	go ds.startDownload(download, settings.APIKey1fichier)

	return download, nil
}

func (ds *downloadService) PauseDownload(id string) (*model.Download, error) {
	if err := ds.workerManager.Pause(id); err != nil {
		return nil, err
	}

	if err := ds.updateStatus(id, model.StatusPaused); err != nil {
		return nil, err
	}

	return ds.downloadRepo.GetByID(id)
}

func (ds *downloadService) ResumeDownload(id string) (*model.Download, error) {
	if err := ds.workerManager.Resume(id); err != nil {
		return nil, err
	}

	if err := ds.updateStatus(id, model.StatusDownloading); err != nil {
		return nil, err
	}

	return ds.downloadRepo.GetByID(id)
}

func (ds *downloadService) CancelDownload(id string) (*model.Download, error) {
	if err := ds.workerManager.Cancel(id); err != nil {
		return nil, err
	}

	if err := ds.updateStatus(id, model.StatusCancelled); err != nil {
		return nil, err
	}

	return ds.downloadRepo.GetByID(id)
}

func (ds *downloadService) ArchiveDownload(id string) error {
	download, err := ds.downloadRepo.GetByID(id)
	if err != nil {
		return err
	}

	download.IsArchived = true
	return ds.downloadRepo.Update(download)
}

func (ds *downloadService) DeleteDownload(id string) error {
	// Annuler si en cours
	ds.workerManager.Cancel(id)
	ds.workerManager.Remove(id)

	download, err := ds.downloadRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Supprimer les fichiers
	if download.DownloadPath != "" && download.Status == model.StatusCompleted {
		os.Remove(download.DownloadPath)
	}

	if download.TempPath != nil && *download.TempPath != "" {
		os.Remove(*download.TempPath)
	}

	return ds.downloadRepo.Delete(id)
}

// Helpers

func (ds *downloadService) startDownload(download *model.Download, apiKey string) {
	// Créer le worker
	w := worker.NewDownloadWorker(download, apiKey)
	ds.workerManager.Add(download.ID, w)
	defer ds.workerManager.Remove(download.ID)

	// Listen for progress updates
	go ds.listenProgress(w)
	// Listen for download infos updates
	go ds.listenInfoReceived(w)

	// Mettre à jour le statut
	ds.updateStatus(download.ID, model.StatusRequesting)

	// Start download
	if err := w.Start(); err != nil {
		// Check if cancelled
		if w.GetState() == worker.StateCancelled {
			ds.updateStatus(download.ID, model.StatusCancelled)
			ds.cleanupTempFile(download)
			return
		}

		// Otherwise, it's an error
		ds.handleDownloadError(download.ID, err.Error())
		ds.cleanupTempFile(download)
		return
	}

	// Download success
	updatedDownload := w.GetDownload()
	updatedDownload.Status = model.StatusCompleted
	ds.downloadRepo.Update(updatedDownload)

	// Send the completion event
	ds.sendCompletedEvent(updatedDownload)
}

func (ds *downloadService) listenProgress(w *worker.DownloadWorker) {
	for update := range w.ProgressChan() {
		download := w.GetDownload()

		// Update database
		ds.downloadRepo.UpdateProgress(
			download.ID,
			update.Progress,
			update.BytesWritten,
			&update.Speed,
			model.StatusDownloading,
		)

		// Send SSE event
		sizeStr := fmt.Sprintf("%d", update.TotalBytes)
		event := &model.DownloadProgressEvent{
			DownloadID:      download.ID,
			FileName:        download.FileName,
			CustomFileName:  download.CustomFileName,
			Status:          string(model.StatusDownloading),
			Progress:        update.Progress,
			DownloadedBytes: fmt.Sprintf("%d", update.BytesWritten),
			FileSize:        &sizeStr,
			Speed:           &update.Speed,
		}

		if err := ds.sseManager.SendEvent("progress", event); err != nil {
			log.Error("Failed to send SSE event:", err)
		}
	}
}

func (ds *downloadService) listenInfoReceived(w *worker.DownloadWorker) {
	for download := range w.InfoReceivedChan() {
		ds.downloadRepo.Update(download)
	}
}

func (ds *downloadService) updateStatus(id string, status model.DownloadStatus) error {
	return ds.downloadRepo.UpdateStatus(id, status)
}

func (ds *downloadService) handleDownloadError(id string, message string) {
	download, _ := ds.downloadRepo.GetByID(id)
	download.ErrorMessage = &message
	download.Status = model.StatusFailed
	ds.downloadRepo.Update(download)

	event := &model.DownloadProgressEvent{
		DownloadID: id,
		Status:     string(model.StatusFailed),
		Progress:   0,
	}

	if err := ds.sseManager.SendEvent("progress", event); err != nil {
		log.Errorf("failed to send error event: %s %v", download.ID, err)
	}
}

func (ds *downloadService) sendCompletedEvent(download *model.Download) {
	if download.FileSize == nil {
		return
	}

	fileSizeStr := fmt.Sprint(*download.FileSize)
	event := &model.DownloadProgressEvent{
		DownloadID:      download.ID,
		FileName:        download.FileName,
		Status:          string(model.StatusCompleted),
		Progress:        100,
		DownloadedBytes: fmt.Sprintf("%d", *download.FileSize),
		FileSize:        &fileSizeStr,
	}

	if err := ds.sseManager.SendEvent("progress", event); err != nil {
		log.Errorf("failed to send complete event: %s %v", download.ID, err)
	}
}

func (ds *downloadService) cleanupTempFile(download *model.Download) error {
	if download.TempPath != nil && *download.TempPath != "" {
		return os.Remove(*download.TempPath)
	}
	return nil
}

func (ds *downloadService) extract1fichierFileID(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL")
	}

	queryString := parsedURL.RawQuery
	fileID := strings.Split(queryString, "&")[0]
	fileID = strings.Split(fileID, "=")[0]

	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validID.MatchString(fileID) {
		return "", fmt.Errorf("id invalid format")
	}

	if len(fileID) > 100 {
		return "", fmt.Errorf("id too long")
	}

	if len(fileID) == 0 {
		return "", fmt.Errorf("id empty")
	}

	return fileID, nil
}
