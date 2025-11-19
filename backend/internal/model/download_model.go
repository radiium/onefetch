package model

import (
	"dlbackend/internal/config"
	"dlbackend/pkg/client"
	"path/filepath"
	"time"
)

type DownloadStatus string

const (
	StatusIdle            DownloadStatus = "IDLE"
	StatusPending         DownloadStatus = "PENDING"
	StatusRequestingInfos DownloadStatus = "REQUESTING_INFOS"
	StatusRequestingToken DownloadStatus = "REQUESTING_TOKEN"
	StatusDownloading     DownloadStatus = "DOWNLOADING"
	StatusPaused          DownloadStatus = "PAUSED"
	StatusCancelled       DownloadStatus = "CANCELLED"
	StatusFailed          DownloadStatus = "FAILED"
	StatusCompleted       DownloadStatus = "COMPLETED"
)

type DownloadType string

const (
	TypeMovie DownloadType = "MOVIE"
	TypeSerie DownloadType = "SERIE"
)

var DownloadTypeDir = map[DownloadType]string{
	TypeMovie: "movies",
	TypeSerie: "series",
}

func (s DownloadType) Dir() string {
	if label, ok := DownloadTypeDir[s]; ok {
		return label
	}
	return "Inconnu"
}

type Download struct {
	ID string `gorm:"primaryKey" json:"id"`

	// Entrées utilisateur
	FileURL        string       `json:"fileUrl"`
	CustomFileDir  *string      `json:"customFileDir"`
	CustomFileName *string      `json:"customFileName"`
	Type           DownloadType `json:"type"`

	// Download infos (depuis api 1fichier.com)
	FileName string  `json:"fileName"`
	FileSize *int64  `json:"fileSize"`
	MimeType *string `json:"mimeType"`
	Checksum *string `json:"checksum"`

	// Download token (depuis api 1fichier.com)
	DownloadURL          *string    `json:"DownloadURL"`
	DownloadURLExpiresAt *time.Time `json:"downloadURLExpiresAt"` // le DirectDownloadURL est valable 5 minutes

	// Gestion status
	Status       DownloadStatus `json:"status"`
	ErrorMessage *string        `json:"errorMessage"`
	StartedAt    *time.Time     `json:"startedAt"`   // Date de démarrage (StatusDownloading)
	CompletedAt  *time.Time     `json:"completedAt"` // Date de fin (StatusCompleted ou StatusFail ou Status)

	// Gestion progression
	Progress        float64  `json:"progress"`
	DownloadedBytes int64    `json:"downloadedBytes"`
	Speed           *float64 `json:"speed"`
	RetryCount      int      `json:"retryCount"`

	// Autres
	CreatedAt  time.Time `json:"createdAt"` // Date de création
	UpdatedAt  time.Time `json:"updatedAt"` // Date de mise à jour
	IsArchived bool      `gorm:"default:false" json:"isArchived"`
}

func (d *Download) resolveFileName() string {
	if d.CustomFileName != nil {
		return filepath.Base(*d.CustomFileName)
	}
	return filepath.Base(d.FileName)
}

func (d *Download) resolveFileDir() (string, error) {
	dirName := ""
	if d.CustomFileDir != nil {
		dirName = *d.CustomFileDir
	}
	fileDir := filepath.Join(config.Cfg.DLPath, d.Type.Dir(), dirName)
	return filepath.Abs(fileDir)
}

func (d *Download) TempFilePath() (string, error) {
	fileName := "." + d.resolveFileName() + ".tmp"
	fileDir, err := d.resolveFileDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(fileDir, fileName), nil
}

func (d *Download) FinalFilePath() (string, error) {
	fileName := d.resolveFileName()
	fileDir, err := d.resolveFileDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(fileDir, fileName), nil
}

func (d *Download) Clone() *Download {
	cp := *d
	return &cp
}

type CreateDownloadRequest struct {
	Type     string  `json:"type"`
	URL      string  `json:"url"`
	FileName *string `json:"fileName"`
	FileDir  *string `json:"fileDir"`
}

type DownloadProgressEvent struct {
	DownloadID      string   `json:"downloadId"`
	FileName        string   `json:"fileName"`
	CustomFileDir   *string  `json:"customFileDir"`
	CustomFileName  *string  `json:"customFileName"`
	Status          string   `json:"status"`
	Progress        float64  `json:"progress"`
	DownloadedBytes string   `json:"downloadedBytes"`
	FileSize        *string  `json:"fileSize"`
	Speed           *float64 `json:"speed"`
}

type DownloadInfoResponse struct {
	Fileinfo    client.OneFichierInfoResponse `json:"fileinfo"`
	Directories map[DownloadType][]string     `json:"directories"`
}
