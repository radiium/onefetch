package service

import (
	"dlbackend/internal/model"
	"dlbackend/internal/repository"
	"dlbackend/pkg/filesystem"
	"dlbackend/pkg/sse"
	"dlbackend/pkg/worker"
	"errors"
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
	CreateDownload(fileURL string, downloadType model.DownloadType) (*model.Download, error)
	PauseDownload(id string) (*model.Download, error)
	ResumeDownload(id string) (*model.Download, error)
	CancelDownload(id string) (*model.Download, error)
	ArchiveDownload(id string) error
	ListDownloads(status []model.DownloadStatus, downloadType []model.DownloadType, page, limit int) ([]model.Download, int64, error)
	DeleteDownload(id string) error
}

type downloadService struct {
	downloadRepo  repository.DownloadRepository
	settingsRepo  repository.SettingsRepository
	sseManager    sse.Manager
	fileManager   filesystem.FileManager
	workerManager *worker.Manager
}

func NewDownloadService(
	downloadRepo repository.DownloadRepository,
	settingsRepo repository.SettingsRepository,
	sseManager sse.Manager,
) DownloadService {
	return &downloadService{
		downloadRepo:  downloadRepo,
		settingsRepo:  settingsRepo,
		sseManager:    sseManager,
		fileManager:   filesystem.NewFileManager(),
		workerManager: worker.NewManager(),
	}
}

func (ds *downloadService) CreateDownload(fileURL string, downloadType model.DownloadType) (*model.Download, error) {
	settings, err := ds.settingsRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	if settings.APIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	// Extraire l'ID du fichier
	fileID, err := ds.extract1fichierFileID(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fileID: %w", err)
	}

	download := &model.Download{
		ID:           uuid.New().String(),
		FileURL:      fileURL,
		FileID:       fileID,
		Type:         downloadType,
		Status:       model.StatusPending,
		Progress:     0,
		DownloadPath: filepath.Join(settings.DownloadPath, downloadType.Dir()),
		RetryCount:   0,
	}

	if err := ds.downloadRepo.Create(download); err != nil {
		return nil, err
	}

	// Démarrer le téléchargement dans une goroutine
	go ds.startDownload(download, settings.APIKey)

	return download, nil
}

func (ds *downloadService) startDownload(download *model.Download, apiKey string) {
	// Créer le worker
	w := worker.NewDownloadWorker(download, apiKey, ds.fileManager)
	ds.workerManager.Add(download.ID, w)

	// Cleanup à la fin
	defer ds.workerManager.Remove(download.ID)

	// Écouter les mises à jour de progression dans une goroutine
	go ds.listenProgress(w)

	// Écouter les mises à jour du download dans une goroutine
	go ds.listenInfoReceived(w)

	// Mettre à jour le statut
	ds.updateStatus(download.ID, model.StatusRequesting)

	// Démarrer le téléchargement
	if err := w.Start(); err != nil {
		// Vérifier si c'était une annulation
		if w.GetState() == worker.StateCancelled {
			ds.updateStatus(download.ID, model.StatusCancelled)
			ds.cleanupTempFile(download)
			return
		}

		// Sinon c'est une erreur
		ds.handleError(download.ID, err.Error())
		ds.cleanupTempFile(download)
		return
	}

	// Téléchargement réussi
	updatedDownload := w.GetDownload()
	updatedDownload.Status = model.StatusCompleted
	ds.downloadRepo.Update(updatedDownload)

	// Envoyer l'événement de complétion
	ds.sendCompletedEvent(updatedDownload)
}

func (ds *downloadService) listenProgress(w *worker.DownloadWorker) {
	for update := range w.ProgressChan() {
		download := w.GetDownload()

		// Mettre à jour la base de données
		ds.downloadRepo.UpdateProgress(
			download.ID,
			update.Progress,
			update.BytesWritten,
			&update.Speed,
		)

		// Envoyer l'événement SSE
		sizeStr := fmt.Sprintf("%d", update.TotalBytes)
		event := &model.DownloadProgressEvent{
			DownloadID:      download.ID,
			FileName:        download.FileName,
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

// Ajouter cette nouvelle fonction
func (ds *downloadService) listenInfoReceived(w *worker.DownloadWorker) {
	for download := range w.InfoReceivedChan() {
		ds.downloadRepo.Update(download)
	}
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

func (ds *downloadService) ListDownloads(status []model.DownloadStatus, downloadTypes []model.DownloadType, page, limit int) ([]model.Download, int64, error) {
	return ds.downloadRepo.List(status, downloadTypes, page, limit)
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

func (ds *downloadService) updateStatus(id string, status model.DownloadStatus) error {
	return ds.downloadRepo.UpdateStatus(id, status)
}

func (ds *downloadService) handleError(id string, message string) {
	ds.downloadRepo.UpdateStatus(id, model.StatusFailed)
	download, _ := ds.downloadRepo.GetByID(id)
	download.ErrorMessage = &message
	ds.downloadRepo.Update(download)

	event := &model.DownloadProgressEvent{
		DownloadID: id,
		Status:     string(model.StatusFailed),
		Progress:   0,
	}

	if err := ds.sseManager.SendEvent("progress", event); err != nil {
		log.Error("Failed to send error event:", err)
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
		log.Error("Failed to send completion event:", err)
	}
}

func (ds *downloadService) cleanupTempFile(download *model.Download) {
	if download.TempPath != nil && *download.TempPath != "" {
		ds.fileManager.RemoveFile(*download.TempPath)
	}
}

func (ds *downloadService) extract1fichierFileID(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", errors.New("URL invalide")
	}

	queryString := parsedURL.RawQuery
	fileID := strings.Split(queryString, "&")[0]
	fileID = strings.Split(fileID, "=")[0]

	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validID.MatchString(fileID) {
		return "", errors.New("format d'ID invalide")
	}

	if len(fileID) > 100 {
		return "", errors.New("ID trop long")
	}

	if len(fileID) == 0 {
		return "", errors.New("ID vide")
	}

	return fileID, nil
}
