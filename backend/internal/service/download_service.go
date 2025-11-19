package service

import (
	"context"
	"dlbackend/internal/config"
	"dlbackend/internal/errors"
	"dlbackend/internal/model"
	"dlbackend/internal/repository"
	"dlbackend/internal/utils"
	"dlbackend/pkg/client"
	"dlbackend/pkg/sse"
	"dlbackend/pkg/worker"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	downloadRepo repository.DownloadRepository
	settingsRepo repository.SettingsRepository
	filesService FilesService
	sseManager   sse.Manager
	dlManager    *worker.DownloadManager
}

func NewDownloadService(
	downloadRepo repository.DownloadRepository,
	settingsRepo repository.SettingsRepository,
	filesService FilesService,
	sseManager sse.Manager,
) DownloadService {
	return &downloadService{
		downloadRepo: downloadRepo,
		settingsRepo: settingsRepo,
		filesService: filesService,
		sseManager:   sseManager,
		dlManager:    worker.NewDownloadManager(context.Background(), downloadRepo, settingsRepo, sseManager),
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
		return nil, fmt.Errorf("get movies directories error: %w", err)
	}

	seriePath := filepath.Join(config.Cfg.DLPath, model.TypeSerie.Dir())
	serieDirectories, err := utils.BuildDirTreeAsList(seriePath)
	if err != nil {
		return nil, fmt.Errorf("get series directories error: %w", err)
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

func (ds *downloadService) CreateDownload(fileURL string, downloadType model.DownloadType, customFileDir string, customFileName string) (*model.Download, error) {
	settings, err := ds.settingsRepo.Get()
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to load settings: %v", err))
	}
	if settings.APIKey1fichier == "" {
		return nil, errors.Internal("1fichier API key not configured")
	}

	// Create Download
	download := &model.Download{
		ID:              uuid.New().String(),
		FileURL:         fileURL,
		CustomFileDir:   &customFileDir,
		CustomFileName:  &customFileName,
		Type:            downloadType,
		Status:          model.StatusPending,
		Progress:        0,
		DownloadedBytes: 0,
		RetryCount:      0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsArchived:      false,
	}

	if err := ds.downloadRepo.Create(download); err != nil {
		return nil, err
	}

	// Start download
	if err := ds.dlManager.Start(download.Clone()); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to Start download: %v", err))
	}

	return download, nil
}

func (ds *downloadService) PauseDownload(id string) (*model.Download, error) {
	if err := ds.dlManager.Pause(id); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to pause download: %v", err))
	}
	return ds.downloadRepo.GetByID(id)
}

func (ds *downloadService) ResumeDownload(id string) (*model.Download, error) {
	if err := ds.dlManager.Resume(id); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to resume download: %v", err))
	}
	return ds.downloadRepo.GetByID(id)
}

func (ds *downloadService) CancelDownload(id string) (*model.Download, error) {
	if err := ds.dlManager.Cancel(id); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to cancel download: %v", err))
	}
	return ds.downloadRepo.GetByID(id)
}

func (ds *downloadService) ArchiveDownload(id string) error {
	download, err := ds.downloadRepo.GetByID(id)
	if err != nil {
		return err
	}
	if download.Status != model.StatusCompleted &&
		download.Status != model.StatusFailed &&
		download.Status != model.StatusCancelled {
		return fmt.Errorf("can't archive an active download. current state %s", download.Status)
	}
	download.IsArchived = true
	return ds.downloadRepo.Update(download)
}

func (ds *downloadService) DeleteDownload(id string) error {
	download, err := ds.downloadRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete final file if exist
	finalPath, err := download.FinalFilePath()
	if finalPath != "" || err != nil {
		return err
	}
	if download.Status == model.StatusCompleted {
		if err := os.Remove(finalPath); err != nil && !os.IsNotExist(err) {
			log.Warnf("Failed to delete file %s: %v", finalPath, err)
		}
	}

	// Delete temp file if exist
	tempPath, err := download.TempFilePath()
	if tempPath != "" && err != nil {
		return err
	}
	if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
		log.Warnf("Failed to delete temp file %s: %v", tempPath, err)
	}

	return ds.downloadRepo.Delete(id)
}

// ============================================================================
// PRIVATE METHODS
// ============================================================================

// cleanupTempFile supprime le fichier temporaire
// func (ds *downloadService) cleanupTempFile(download *model.Download) {
// 	if download.TempPath != nil && *download.TempPath != "" {
// 		if err := os.Remove(*download.TempPath); err != nil && !os.IsNotExist(err) {
// 			log.Warnf("Failed to cleanup temp file %s: %v", *download.TempPath, err)
// 		}
// 	}
// }

// extract1fichierFileID extrait l'ID du fichier depuis l'URL
// func (ds *downloadService) extract1fichierFileID(rawURL string) (string, error) {
// 	parsedURL, err := url.Parse(rawURL)
// 	if err != nil {
// 		return "", fmt.Errorf("invalid URL")
// 	}

// 	queryString := parsedURL.RawQuery
// 	if queryString == "" {
// 		return "", fmt.Errorf("no file ID in URL")
// 	}

// 	firstParam := strings.Split(queryString, "&")[0]
// 	if strings.Contains(firstParam, "=") {
// 		return "", fmt.Errorf("invalid URL format: expected /?fileID")
// 	}

// 	fileID := firstParam

// 	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
// 	if !validID.MatchString(fileID) {
// 		return "", fmt.Errorf("invalid ID format")
// 	}

// 	if len(fileID) > 100 {
// 		return "", fmt.Errorf("ID too long")
// 	}

// 	if len(fileID) == 0 {
// 		return "", fmt.Errorf("ID empty")
// 	}

// 	return fileID, nil
// }
