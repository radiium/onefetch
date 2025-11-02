package worker

import (
	"context"
	"dlbackend/internal/model"
	"dlbackend/pkg/client"
	"dlbackend/pkg/filesystem"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// DownloadWorker gère un téléchargement individuel avec son propre cycle de vie
type DownloadWorker struct {
	download         *model.Download
	client           client.OneFichierClient
	fileManager      filesystem.FileManager
	ctx              context.Context
	cancel           context.CancelFunc
	mu               sync.RWMutex
	state            WorkerState
	pauseChan        chan struct{}
	resumeChan       chan struct{}
	progressChan     chan *ProgressUpdate
	infoReceivedChan chan *model.Download
}

type WorkerState int

const (
	StateRunning WorkerState = iota
	StatePaused
	StateCancelled
	StateCompleted
	StateFailed
)

type ProgressUpdate struct {
	BytesWritten int64
	TotalBytes   int64
	Speed        float64
	Progress     float64
}

// NewDownloadWorker crée un nouveau worker pour un téléchargement
func NewDownloadWorker(
	download *model.Download,
	apiKey string,
	fileManager filesystem.FileManager,
) *DownloadWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &DownloadWorker{
		download:         download,
		client:           client.NewOneFichierClient(apiKey),
		fileManager:      fileManager,
		ctx:              ctx,
		cancel:           cancel,
		state:            StateRunning,
		pauseChan:        make(chan struct{}),
		resumeChan:       make(chan struct{}),
		progressChan:     make(chan *ProgressUpdate, 10),
		infoReceivedChan: make(chan *model.Download, 10),
	}
}

// Start démarre le téléchargement
func (w *DownloadWorker) Start() error {
	// Étape 1: Récupérer les infos du fichier
	if err := w.fetchFileInfo(); err != nil {
		return fmt.Errorf("failed to fetch file info: %w", err)
	}

	// Étape 2: Obtenir le token de téléchargement
	if err := w.fetchDownloadToken(); err != nil {
		return fmt.Errorf("failed to fetch download token: %w", err)
	}

	// Étape 3: Télécharger le fichier
	if err := w.downloadFile(); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	w.setState(StateCompleted)
	return nil
}

// fetchFileInfo récupère les informations du fichier
func (w *DownloadWorker) fetchFileInfo() error {
	info, err := w.client.GetFileInfo(w.download.FileURL)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.download.FileName = info.Filename
	w.download.FileSize = &info.Size
	w.download.MimeType = &info.ContentType
	w.download.Checksum = &info.Checksum
	w.mu.Unlock()

	// Envoyer la mise à jour dans le channel
	select {
	case w.infoReceivedChan <- w.download:
	default:
	}

	return nil
}

// fetchDownloadToken obtient l'URL de téléchargement direct
func (w *DownloadWorker) fetchDownloadToken() error {
	token, err := w.client.GetDownloadToken(w.download.FileURL)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.download.DirectDownloadURL = &token.URL
	expiresAt := time.Now().Add(5 * time.Minute)
	w.download.DirectURLExpiresAt = &expiresAt
	w.mu.Unlock()

	return nil
}

// Dans download_worker.go, modifier downloadFile
func (w *DownloadWorker) downloadFile() error {
	if err := w.fileManager.EnsureDir(w.download.DownloadPath); err != nil {
		return err
	}

	fileName := w.download.FileName
	if w.download.CustomFileName != nil && *w.download.CustomFileName != "" {
		fileName = *w.download.CustomFileName
	}

	tempPath := filepath.Join(w.download.DownloadPath, "."+fileName+".tmp")
	finalPath := filepath.Join(w.download.DownloadPath, fileName)

	w.mu.Lock()
	w.download.TempPath = &tempPath
	w.mu.Unlock()

	src, fileSize, err := w.client.DownloadFile(*w.download.DirectDownloadURL)
	if err != nil {
		return err
	}
	defer src.Close()

	if fileSize <= 0 {
		fileSize = 1
	}

	// Wrapper pour gérer la pause/reprise dans le callback
	progressCallback := w.createProgressCallback(fileSize)

	// Télécharger avec gestion du contexte
	if err := w.fileManager.WriteTempFileWithContext(w.ctx, src, tempPath, progressCallback); err != nil {
		w.fileManager.RemoveFile(tempPath)
		return err
	}

	// Déplacer le fichier temporaire vers le fichier final
	if err := w.fileManager.MoveFile(tempPath, finalPath); err != nil {
		return err
	}

	w.mu.Lock()
	w.download.Progress = 100
	w.download.DownloadedBytes = fileSize
	now := time.Now()
	w.download.CompletedAt = &now
	w.mu.Unlock()

	return nil
}

// createProgressCallback crée un callback qui gère la pause/reprise
func (w *DownloadWorker) createProgressCallback(totalBytes int64) func(int64) {
	startTime := time.Now()
	lastReportTime := startTime
	var lastBytesWritten int64

	return func(bytesWritten int64) {
		// Gérer la pause
		w.waitIfPaused()

		// Calculer la progression
		progress := float64(bytesWritten) / float64(totalBytes) * 100
		elapsed := time.Since(startTime).Seconds()
		var speed float64
		if elapsed > 0 {
			speed = float64(bytesWritten) / elapsed
		}

		// Throttle les updates (max toutes les 500ms)
		now := time.Now()
		if now.Sub(lastReportTime) < 500*time.Millisecond && bytesWritten < totalBytes {
			return
		}
		lastReportTime = now

		if bytesWritten > lastBytesWritten {
			lastBytesWritten = bytesWritten

			// Envoyer l'update de progression
			select {
			case w.progressChan <- &ProgressUpdate{
				BytesWritten: bytesWritten,
				TotalBytes:   totalBytes,
				Speed:        speed,
				Progress:     progress,
			}:
			default:
				// Channel plein, skip cet update
			}
		}
	}
}

// waitIfPaused bloque tant que le worker est en pause
func (w *DownloadWorker) waitIfPaused() {
	w.mu.RLock()
	state := w.state
	w.mu.RUnlock()

	if state == StatePaused {
		select {
		case <-w.resumeChan:
			return
		case <-w.ctx.Done():
			return
		}
	}
}

// Pause met le téléchargement en pause
func (w *DownloadWorker) Pause() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state != StateRunning {
		return fmt.Errorf("cannot pause: worker not running")
	}

	w.state = StatePaused
	return nil
}

// Resume reprend le téléchargement
func (w *DownloadWorker) Resume() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state != StatePaused {
		return fmt.Errorf("cannot resume: worker not paused")
	}

	w.state = StateRunning

	// Débloquer tous les goroutines en attente
	select {
	case w.resumeChan <- struct{}{}:
	default:
	}

	return nil
}

// Cancel annule le téléchargement
func (w *DownloadWorker) Cancel() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state == StateCancelled || w.state == StateCompleted {
		return fmt.Errorf("cannot cancel: download already finished")
	}

	w.state = StateCancelled
	w.cancel() // Annule le contexte
	return nil
}

// GetState retourne l'état actuel du worker
func (w *DownloadWorker) GetState() WorkerState {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state
}

// setState change l'état du worker
func (w *DownloadWorker) setState(state WorkerState) {
	w.mu.Lock()
	w.state = state
	w.mu.Unlock()
}

// GetDownload retourne le modèle de téléchargement
func (w *DownloadWorker) GetDownload() *model.Download {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.download
}

// ProgressChan retourne le channel de progression
func (w *DownloadWorker) ProgressChan() <-chan *ProgressUpdate {
	return w.progressChan
}

// UpdateChan retourne le channel d'update
func (w *DownloadWorker) InfoReceivedChan() <-chan *model.Download {
	return w.infoReceivedChan
}

// Close nettoie les ressources du worker
func (w *DownloadWorker) Close() {
	w.cancel()
	close(w.progressChan)
	close(w.infoReceivedChan)
}
