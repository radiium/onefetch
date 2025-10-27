## Header

- Title => DL Manager
- Version => 1

## Tags

### Settings

- Description => App settings

### Downloads

- Description => Downloads made in the application

## Operations

### GET /api/settings

- Tag => Settings
- Description => Get settings
- Request => empty
- Response 200 => `Settings`

### PATCH /api/settings

- Tag => Settings
- Description => Update settings
- Request => `Settings`
- Response 200 => `Settings`

### POST /api/downloads

- Tag => Downloads
- Description => Create and start a download
- Request (body) => `CreateDownloadRequest`
- Response 201 => `Download`

### GET /api/downloads

- Tag => Downloads
- Description => List of downloads with pagination, filters and sorting
- Request (query params) => `DownloadQueryParams`
- Response 200 => `DownloadPage`

### GET /api/downloads/streams

- Tag => Downloads
- Description => Server sent event of active downloads
- Request => empty
- Response 200 => `[]DownloadProgressEvent`

### GET /api/downloads/:id/pause

- Tag => Downloads
- Description => Pause a download
- Request (pathParam) => `Download.ID`
- Response 200 => empty

### GET /api/downloads/:id/resume.

- Tag => Downloads
- Description => Resume a download
- Request (pathParam) => `Download.ID`
- Response 200 => empty

### GET /api/downloads/:id/archive

- Tag => Downloads
- Description => Archive a download
- Request (pathParam) => `Download.ID`
- Response 200 => empty



## Schemas

### Enums

```go
// DownloadStatus représente le statut de téléchargement
type DownloadStatus string

const (
	StatusPending DownloadStatus = "PENDING"
	StatusRequesting DownloadStatus = "REQUESTING"
	StatusDownloading DownloadStatus = "DOWNLOADING"
	StatusPaused DownloadStatus = "PAUSED"
	StatusCompleted DownloadStatus = "COMPLETED"
	StatusFailed DownloadStatus = "FAILED"
	StatusCancelled DownloadStatus = "CANCELLED"
)

// DownloadType représente le type de contenu téléchargé
type DownloadType string

const (
	TypeMovie DownloadType = "MOVIE"
	TypeTVShow DownloadType = "TVSHOW"
)
```

### DTOs

```go
type CreateDownloadRequest struct {
	URL []string `json:"url"`
	Type DownloadType `json:"type"`
}

type DownloadProgressEvent struct {
	DownloadID string `json:"downloadId"`
	FileName string `json:"fileName"`
	Status DownloadStatus `json:"status"`
	Progress float64 `json:"progress"`
	DownloadedBytes *big.Int `json:"downloadedBytes"`
	FileSize *big.Int `json:"fileSize,omitempty"`
	Speed *float64 `json:"speed,omitempty"`
}

type DownloadQueryParams struct {
	Status *DownloadStatus `json:"status,omitempty"`
	Type *DownloadType `json:"type,omitempty"`
	Page *int `json:"page,omitempty"`
	Limit *int `json:"limit,omitempty"`
}

type Pagination struct {
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type DownloadPage struct {
	Data []Download `json:"data"`
	Pagination Pagination `json:"pagination"`
}
```

### Entities

```go
type Settings struct {
	APIKey string `json:"apiKey"`
	DownloadPath string `json:"downloadPath"`
}

type Download struct {
	ID string `json:"id"`
	FileURL string `json:"fileUrl"`
	FileID string `json:"fileId"`
	FileName string `json:"fileName"`
	FileSize *int64 `json:"fileSize,omitempty"`
	MimeType *string `json:"mimeType,omitempty"`
	Checksum *string `json:"checksum,omitempty"`
	Type DownloadType `json:"type"`
	DirectDownloadURL *string `json:"directDownloadUrl,omitempty"`
	DirectURLExpiresAt *time.Time `json:"directUrlExpiresAt,omitempty"`
	Status DownloadStatus `json:"status"`
	Progress float64 `json:"progress"`
	DownloadedBytes int64 `json:"downloadedBytes"`
	Speed *float64 `json:"speed,omitempty"`
	DownloadPath string `json:"downloadPath"`
	TempPath *string `json:"tempPath,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
	RetryCount int `json:"retryCount"`
}
```

## L'API 1fichier.com

Le téléchargement des fichiers par

Le téléchargement des fichiers ce fait via l'api 1fichier avec apiKey et compte premium et en 3 étapes:

### étape 1

Récupération des infos du fichier a télécharger

#### Endpoint:

POST https://api.1fichier.com/v1/file/info.cgi + Header 'Authorization: Bearer {Settings.APIKey}

#### Réponse

```go
type OneFichierInfoResponse struct {
	URL string `json:"url"`
	Filename string `json:"filename"`
	Size int64 `json:"size"`
	Date time.Time `json:"date"`
	Checksum string `json:"checksum"`
	ContentType string `json:"content_type"`
	Description \*string `json:"description,omitempty"`
	Pass int `json:"pass"` // 0 ou 1
	Path string `json:"path"`
	FolderID string `json:"folder_id"`
}
```

### étape 2

Récupération du lien de téléchargement final du fichier

#### Endpoint:

POST https://api.1fichier.com/v1/download/get_token.cgi + Header 'Authorization: Bearer {Settings.APIKey}

#### Réponse

```go
type OneFichierTokenResponse struct {
	URL string `json:"url"`
	Status string `json:"status"` // "OK" ou "KO"
	Message \*string `json:"message,omitempty"`
}
```

### étape 3

Création le l'entrée Download et démarrage du téléchargement
