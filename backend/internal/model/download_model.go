package model

import "time"

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
	TypeMovie  DownloadType = "MOVIE"
	TypeTVShow DownloadType = "TVSHOW"
)

var DownloadTypeDir = map[DownloadType]string{
	TypeMovie:  "movies",
	TypeTVShow: "tvshows",
}

func (s DownloadType) Dir() string {
	if label, ok := DownloadTypeDir[s]; ok {
		return label
	}
	return "Inconnu"
}

type Download struct {
	ID                 string         `gorm:"primaryKey" json:"id"`
	FileURL            string         `json:"fileUrl"`
	FileID             string         `json:"fileId"`
	FileName           string         `json:"fileName"`
	FileSize           *int64         `json:"fileSize"`
	MimeType           *string        `json:"mimeType"`
	Checksum           *string        `json:"checksum"`
	Type               DownloadType   `json:"type"`
	DirectDownloadURL  *string        `json:"directDownloadUrl"`
	DirectURLExpiresAt *time.Time     `json:"directUrlExpiresAt"`
	Status             DownloadStatus `json:"status"`
	Progress           float64        `json:"progress"`
	DownloadedBytes    int64          `json:"downloadedBytes"`
	Speed              *float64       `json:"speed"`
	DownloadPath       string         `json:"downloadPath"`
	TempPath           *string        `json:"tempPath"`
	CreatedAt          time.Time      `json:"createdAt"`
	StartedAt          *time.Time     `json:"startedAt"`
	CompletedAt        *time.Time     `json:"completedAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	ErrorMessage       *string        `json:"errorMessage"`
	RetryCount         int            `json:"retryCount"`
	IsArchived         bool           `gorm:"default:false"`
}

type CreateDownloadRequest struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type DownloadProgressEvent struct {
	DownloadID      string   `json:"downloadId"`
	FileName        string   `json:"fileName"`
	Status          string   `json:"status"`
	Progress        float64  `json:"progress"`
	DownloadedBytes string   `json:"downloadedBytes"`
	FileSize        *string  `json:"fileSize"`
	Speed           *float64 `json:"speed"`
}
