## Contexte

Tu développe un gestionnaire de téléchargement spécial pour des liens 1fichier.com via API avec compte premium.
tu as déjà un frontend svelteKit contenant 2 pages:

- La page /downloads affiche la liste de tous les téléchargements effectués.
- La page /home contient un formulaire pour envoyer un lien de téléchargement et afficher les téléchargements en cours.

## Objectif

Tu dois développer un backend en go avec Fiber, Sqlite.
En évitant la complexité inutile et en utilisant une structure de dossier standard.

## .env

Implémenter un fichier .env avec les variables suivantes:

- APP_HOST
- APP_PORT
- APP_DOWNLOAD_PATH

## Endpoints:

Implémenter les endpoints suivants:

- POST /api/downloads => Creation et démarrage dun download
- GET /api/downloads => liste des downloads avec filtres, trie et pagination
- GET /api/downloads/streams => Server sent Event des downloads actifs / en cours
- GET /api/downloads/:id/pause => Pause un download
- GET /api/downloads/:id/resume. => Reprend un downloads
- GET /api/downloads/:id/archive => Archive un downloads
- GET /api/settings => Get settings
- PUT /api/settings => Update settings

## Les DTOs et entité à implémenter

### DTO CreateDownloadRequest:

Utilisé pour la création de Download (POST /api/downloads)

```go
type CreateDownloadRequest struct {
	URL []string `json:"url"`
	Type DownloadType `json:"type"`
}
```

### DTO DownloadProgressEvent:

- Utilisé pour le stream SSE (GET /api/downloads/streams)

```go
type DownloadProgressEvent struct {
	DownloadID string `json:"downloadId"`
	FileName string `json:"fileName"`
	Status DownloadStatus `json:"status"`
	Progress float64 `json:"progress"`
	DownloadedBytes *big.Int `json:"downloadedBytes"`
	FileSize *big.Int `json:"fileSize,omitempty"`
	Speed \*float64 `json:"speed,omitempty"`
}
```

### DTO DownloadQueryParams:

Query params utilisé pour récupérer la liste des downloads (GET /api/downloads)

```go
type DownloadQueryParams struct {
	Status *DownloadStatus `json:"status,omitempty"`
	Type *DownloadType `json:"type,omitempty"`
	Page *int `json:"page,omitempty"`
	Limit *int `json:"limit,omitempty"`
}
```

### DTO DownloadPage et Pagination:

Retour de la liste des downloads (GET /api/downloads)

```go
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

### Entité Settings:

Représente les paramètres de l'application en base. C'est une entrée unique en base

```go
type Settings struct {
	APIKey string `json:"apiKey"`
	DownloadPath string `json:"downloadPath"`
}
```

### Entité Download

Représente un téléchargement en base

```go
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

### Enums:

```go
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

## L'API 1fichier.com

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
