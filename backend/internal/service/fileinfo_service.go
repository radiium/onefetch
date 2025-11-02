package service

import (
	"dlbackend/internal/model"
	"dlbackend/internal/repository"
	"dlbackend/pkg/client"
	"dlbackend/pkg/filesystem"
	"fmt"
	"path/filepath"
)

type FileinfoService interface {
	GetFileinfo(fileURL string) (*model.FileinfoResponse, error)
}

type fileinfoService struct {
	settingsRepo repository.SettingsRepository
	fileManager  filesystem.FileManager
}

func NewFileinfoService(
	settingsRepo repository.SettingsRepository,
) FileinfoService {
	return &fileinfoService{
		settingsRepo: settingsRepo,
		fileManager:  filesystem.NewFileManager(),
	}
}

func (fs *fileinfoService) GetFileinfo(fileURL string) (*model.FileinfoResponse, error) {
	settings, err := fs.settingsRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("Failed to get settings: %w", err)
	}

	if settings.APIKey1fichier == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	oneFichierClient := client.NewOneFichierClient(settings.APIKey1fichier)
	fileinfo, err := oneFichierClient.GetFileInfo(fileURL)
	if err != nil {
		return nil, fmt.Errorf("1fichier api error: %w", err)
	}

	moviePath := filepath.Join(settings.DownloadPath, model.TypeMovie.Dir())
	movieDirectories, err := fs.fileManager.GetDirectories(moviePath)
	if err != nil {
		return nil, fmt.Errorf("Get movies directories error: %w", err)
	}

	seriePath := filepath.Join(settings.DownloadPath, model.TypeSerie.Dir())
	serieDirectories, err := fs.fileManager.GetDirectories(seriePath)
	if err != nil {
		return nil, fmt.Errorf("Get series directories error: %w", err)
	}

	return &model.FileinfoResponse{
		Fileinfo: *fileinfo,
		Directories: map[model.DownloadType][]string{
			model.TypeMovie: movieDirectories,
			model.TypeSerie: serieDirectories,
		},
	}, nil
}
