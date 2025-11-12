package model

import (
	"dlbackend/pkg/client"
	"time"
)

type DownloadStatus string

const (
	StatusPending     DownloadStatus = "PENDING"
	StatusRequesting  DownloadStatus = "REQUESTING"
	StatusDownloading DownloadStatus = "DOWNLOADING"
	StatusPaused      DownloadStatus = "PAUSED"
	StatusCompleted   DownloadStatus = "COMPLETED"
	StatusFailed      DownloadStatus = "FAILED"
	StatusCancelled   DownloadStatus = "CANCELLED"
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
	ID             string       `gorm:"primaryKey" json:"id"`
	FileURL        string       `json:"fileUrl"`
	FileID         string       `json:"fileId"`
	FileName       string       `json:"fileName"`
	CustomFileName *string      `json:"customFileName"`
	FileSize       *int64       `json:"fileSize"`
	MimeType       *string      `json:"mimeType"`
	Checksum       *string      `json:"checksum"`
	Type           DownloadType `json:"type"`

	DirectDownloadURL  *string    `json:"directDownloadUrl"`
	DirectURLExpiresAt *time.Time `json:"directUrlExpiresAt"`

	DownloadPath string  `json:"downloadPath"`
	TempPath     *string `json:"tempPath"`

	Status          DownloadStatus `json:"status"`
	Progress        float64        `json:"progress"`
	DownloadedBytes int64          `json:"downloadedBytes"`
	Speed           *float64       `json:"speed"`
	ErrorMessage    *string        `json:"errorMessage"`
	RetryCount      int            `json:"retryCount"`

	CreatedAt   time.Time  `json:"createdAt"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	IsArchived bool `gorm:"default:false" json:"isArchived"`
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
