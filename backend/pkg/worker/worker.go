package worker

import (
	"context"
	"dlbackend/internal/model"
	"dlbackend/internal/repository"
	"dlbackend/pkg/client"
	"dlbackend/pkg/sse"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

// ============================================================================
// DOWNLOAD MANAGER
// ============================================================================

type DownloadManager struct {
	workers      sync.Map
	repo         repository.DownloadRepository
	settingsRepo repository.SettingsRepository
	sseManager   sse.Manager
	ctx          context.Context
}

func NewDownloadManager(
	ctx context.Context,
	repo repository.DownloadRepository,
	settingsRepo repository.SettingsRepository,
	sseManager sse.Manager,
) *DownloadManager {
	return &DownloadManager{
		ctx:          ctx,
		repo:         repo,
		settingsRepo: settingsRepo,
		sseManager:   sseManager,
	}
}

func (m *DownloadManager) Start(download *model.Download) error {
	settings, err := m.settingsRepo.Get()
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}
	if settings.APIKey1fichier == "" {
		return errors.New("API key not configured")
	}

	client := client.NewOneFichierClient(settings.APIKey1fichier)
	worker := NewDownloadWorker(m.ctx, download, m.repo, client, m.sseManager)

	m.workers.Store(download.ID, worker)

	go func() {
		defer m.workers.Delete(download.ID)
		worker.Run()
	}()

	return nil
}

func (m *DownloadManager) Pause(downloadID string) error {
	value, ok := m.workers.Load(downloadID)
	if !ok {
		return errors.New("download not found")
	}

	worker := value.(*DownloadWorker)
	worker.Pause()
	return nil
}

func (m *DownloadManager) Resume(downloadID string) error {
	value, ok := m.workers.Load(downloadID)
	if !ok {
		return errors.New("download not found")
	}

	worker := value.(*DownloadWorker)
	worker.Resume()
	return nil
}

func (m *DownloadManager) Cancel(downloadID string) error {
	value, ok := m.workers.Load(downloadID)
	if !ok {
		return errors.New("download not found")
	}

	worker := value.(*DownloadWorker)
	worker.Cancel()
	return nil
}

// ============================================================================
// DOWNLOAD WORKER - Architecture Event-Driven
// ============================================================================

type DownloadWorker struct {
	download   *model.Download
	repo       repository.DownloadRepository
	client     client.OneFichierClient
	sseManager sse.Manager

	// Contrôle via atomic (pas de mutex nécessaire!)
	state atomic.Int32 // 0=running, 1=paused, 2=cancelled

	ctx    context.Context
	cancel context.CancelFunc

	// File writing
	file *os.File
	mu   sync.Mutex // Uniquement pour les opérations fichier
}

const (
	StateRunning = iota
	StatePaused
	StateCancelled
)

func NewDownloadWorker(
	ctx context.Context,
	download *model.Download,
	repo repository.DownloadRepository,
	client client.OneFichierClient,
	sseManager sse.Manager,
) *DownloadWorker {
	workerCtx, cancel := context.WithCancel(ctx)

	w := &DownloadWorker{
		download:   download,
		repo:       repo,
		client:     client,
		sseManager: sseManager,
		ctx:        workerCtx,
		cancel:     cancel,
	}

	w.state.Store(StateRunning)
	return w
}

// Pause demande la pause (thread-safe, non-bloquant)
func (w *DownloadWorker) Pause() {
	w.state.Store(StatePaused)
	log.Infof("Pause requested for download %s", w.download.ID)
}

// Resume annule la pause (thread-safe, non-bloquant)
func (w *DownloadWorker) Resume() {
	if w.state.CompareAndSwap(StatePaused, StateRunning) {
		log.Infof("Resume requested for download %s", w.download.ID)
	}
}

// Cancel annule le téléchargement (idempotent)
func (w *DownloadWorker) Cancel() {
	if w.state.Swap(StateCancelled) != StateCancelled {
		log.Infof("Cancel requested for download %s", w.download.ID)
		w.cancel()
	}
}

// IsPaused vérifie si en pause
func (w *DownloadWorker) IsPaused() bool {
	return w.state.Load() == StatePaused
}

// IsCancelled vérifie si annulé
func (w *DownloadWorker) IsCancelled() bool {
	return w.state.Load() == StateCancelled
}

// UpdateDownload met à jour les données (thread-safe)
func (w *DownloadWorker) UpdateDownload(fn func(*model.Download)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fn(w.download)
}

// notifyProgress sauvegarde et envoie SSE
func (w *DownloadWorker) notifyProgress() {
	if err := w.repo.Update(w.download); err != nil {
		log.Errorf("Failed to update DB for download %s: %v", w.download.ID, err)
	}

	event := model.DownloadProgressEvent{
		DownloadID:      w.download.ID,
		FileName:        w.download.FileName,
		CustomFileDir:   w.download.CustomFileDir,
		CustomFileName:  w.download.CustomFileName,
		Status:          string(w.download.Status),
		Progress:        w.download.Progress,
		DownloadedBytes: fmt.Sprintf("%d", w.download.DownloadedBytes),
		Speed:           w.download.Speed,
	}
	if w.download.FileSize != nil {
		size := fmt.Sprintf("%d", *w.download.FileSize)
		event.FileSize = &size
	}

	if err := w.sseManager.SendEvent("progress", event); err != nil {
		log.Errorf("Failed to send SSE for download %s: %v", w.download.ID, err)
	}
}

// Run exécute le workflow complet
func (w *DownloadWorker) Run() error {
	defer w.cleanup()

	// Étapes séquentielles
	if err := w.stepGetFileInfo(); err != nil {
		return w.fail(err)
	}

	if err := w.stepGetDownloadToken(); err != nil {
		return w.fail(err)
	}

	if err := w.stepDownload(); err != nil {
		if w.IsCancelled() {
			return w.cancelCleanup()
		}
		return w.fail(err)
	}

	return w.complete()
}

// stepGetFileInfo récupère les infos du fichier
func (w *DownloadWorker) stepGetFileInfo() error {
	w.UpdateDownload(func(d *model.Download) {
		d.Status = model.StatusRequestingInfos
	})
	w.notifyProgress()

	info, err := w.client.GetFileInfo(w.download.FileURL)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	w.UpdateDownload(func(d *model.Download) {
		d.FileName = info.Filename
		d.FileSize = &info.Size
		d.Checksum = &info.Checksum
		if info.ContentType != "" {
			d.MimeType = &info.ContentType
		}
	})
	w.notifyProgress()

	return nil
}

// stepGetDownloadToken récupère le token de téléchargement
func (w *DownloadWorker) stepGetDownloadToken() error {
	w.UpdateDownload(func(d *model.Download) {
		d.Status = model.StatusRequestingToken
	})
	w.notifyProgress()

	token, err := w.client.GetDownloadToken(w.download.FileURL)
	if err != nil {
		return fmt.Errorf("failed to get download token: %w", err)
	}

	w.UpdateDownload(func(d *model.Download) {
		d.DownloadURL = &token.URL
		expiresAt := time.Now().Add(5 * time.Minute)
		d.DownloadURLExpiresAt = &expiresAt
	})
	w.notifyProgress()

	return nil
}

// stepDownload effectue le téléchargement avec gestion pause/resume
func (w *DownloadWorker) stepDownload() error {
	now := time.Now()
	w.UpdateDownload(func(d *model.Download) {
		d.Status = model.StatusDownloading
		if d.StartedAt == nil {
			d.StartedAt = &now
		}
	})
	w.notifyProgress()

	if w.download.DownloadURL == nil {
		return errors.New("no download URL")
	}

	// Préparer le fichier
	if err := w.prepareFile(); err != nil {
		return err
	}
	defer w.closeFile()

	// Boucle de téléchargement avec reprise automatique après pause
	for {
		if w.IsCancelled() {
			return errors.New("cancelled")
		}

		// Si en pause, attendre
		if w.IsPaused() {
			w.UpdateDownload(func(d *model.Download) {
				d.Status = model.StatusPaused
			})
			w.notifyProgress()

			log.Debugf("Download %s paused, waiting for resume...", w.download.ID)

			// Attente active avec vérification périodique
			for w.IsPaused() && !w.IsCancelled() {
				time.Sleep(100 * time.Millisecond)
			}

			if w.IsCancelled() {
				return errors.New("cancelled")
			}

			// Reprendre
			w.UpdateDownload(func(d *model.Download) {
				d.Status = model.StatusDownloading
			})
			w.notifyProgress()
			log.Debugf("Download %s resumed", w.download.ID)
		}

		// Télécharger un chunk
		completed, err := w.downloadChunk()
		if err != nil {
			return err
		}

		if completed {
			return nil
		}
	}
}

// prepareFile prépare le fichier pour l'écriture
func (w *DownloadWorker) prepareFile() error {
	tempPath, err := w.download.TempFilePath()
	if err != nil {
		return fmt.Errorf("failed to resolve temp path: %w", err)
	}

	// Créer le répertoire
	if err := os.MkdirAll(filepath.Dir(tempPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Vérifier si fichier existe pour reprise
	if w.download.DownloadedBytes > 0 {
		stat, err := os.Stat(tempPath)
		if err != nil || stat.Size() != w.download.DownloadedBytes {
			log.Warnf("Temp file mismatch, restarting from zero: %s", w.download.ID)
			w.UpdateDownload(func(d *model.Download) {
				d.DownloadedBytes = 0
			})
		}
	}

	// Ouvrir/créer le fichier
	if w.download.DownloadedBytes > 0 {
		w.file, err = os.OpenFile(tempPath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		w.file, err = os.Create(tempPath)
	}

	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}

	return nil
}

// downloadChunk télécharge un chunk de données
func (w *DownloadWorker) downloadChunk() (completed bool, err error) {
	reader, contentLength, statusCode, err := w.client.DownloadFile(
		*w.download.DownloadURL,
		w.download.DownloadedBytes,
	)
	if err != nil {
		return false, fmt.Errorf("failed to start download: %w", err)
	}
	defer reader.Close()

	// Calculer la taille totale
	totalSize := w.calculateTotalSize(statusCode, contentLength)

	// Lire et écrire avec vérifications régulières
	buffer := make([]byte, 64*1024)
	lastUpdate := time.Now()
	lastBytes := w.download.DownloadedBytes
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Vérifier état
		if w.IsCancelled() {
			return false, errors.New("cancelled")
		}

		if w.IsPaused() {
			log.Debugf("Pause detected during download chunk for %s", w.download.ID)
			return false, nil // Retour sans erreur, la boucle principale gérera la pause
		}

		// Mettre à jour la vitesse périodiquement
		select {
		case <-ticker.C:
			w.updateSpeed(&lastUpdate, &lastBytes)
		default:
		}

		// Lire
		n, readErr := reader.Read(buffer)

		if n > 0 {
			// Écrire
			if _, err := w.file.Write(buffer[:n]); err != nil {
				return false, fmt.Errorf("failed to write: %w", err)
			}

			// Mettre à jour progression
			w.UpdateDownload(func(d *model.Download) {
				d.DownloadedBytes += int64(n)
				if totalSize > 0 {
					d.Progress = float64(d.DownloadedBytes) / float64(totalSize) * 100
				}
			})
		}

		if readErr == io.EOF {
			w.updateSpeed(&lastUpdate, &lastBytes)
			return true, nil // Téléchargement terminé
		}

		if readErr != nil {
			return false, readErr
		}
	}
}

// calculateTotalSize calcule la taille totale
func (w *DownloadWorker) calculateTotalSize(statusCode int, contentLength int64) int64 {
	if w.download.FileSize != nil && *w.download.FileSize > 0 {
		return *w.download.FileSize
	}

	var totalSize int64
	if statusCode == http.StatusPartialContent {
		totalSize = w.download.DownloadedBytes + contentLength
	} else {
		totalSize = contentLength
		if w.download.DownloadedBytes > 0 {
			log.Warnf("Server rejected resume for %s", w.download.ID)
			w.UpdateDownload(func(d *model.Download) {
				d.DownloadedBytes = 0
			})
		}
	}

	w.UpdateDownload(func(d *model.Download) {
		d.FileSize = &totalSize
	})

	return totalSize
}

// updateSpeed met à jour la vitesse
func (w *DownloadWorker) updateSpeed(lastUpdate *time.Time, lastBytes *int64) {
	duration := time.Since(*lastUpdate).Seconds()
	if duration > 0 {
		currentBytes := w.download.DownloadedBytes
		speed := float64(currentBytes-*lastBytes) / duration
		w.UpdateDownload(func(d *model.Download) {
			d.Speed = &speed
		})
		w.notifyProgress()
		*lastUpdate = time.Now()
		*lastBytes = currentBytes
	}
}

// closeFile ferme le fichier
func (w *DownloadWorker) closeFile() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		w.file.Sync()
		w.file.Close()
		w.file = nil
	}
}

// complete marque comme terminé
func (w *DownloadWorker) complete() error {
	tempPath, _ := w.download.TempFilePath()
	finalPath, _ := w.download.FinalFilePath()

	// Supprimer le fichier final s'il existe
	os.Remove(finalPath)

	// Renommer
	if err := os.Rename(tempPath, finalPath); err != nil {
		return fmt.Errorf("failed to finalize: %w", err)
	}

	now := time.Now()
	w.UpdateDownload(func(d *model.Download) {
		d.Status = model.StatusCompleted
		d.Progress = 100
		d.CompletedAt = &now
	})
	w.notifyProgress()

	log.Infof("Download %s completed", w.download.ID)
	return nil
}

// fail marque comme échoué
func (w *DownloadWorker) fail(err error) error {
	errMsg := err.Error()
	w.UpdateDownload(func(d *model.Download) {
		d.Status = model.StatusFailed
		d.ErrorMessage = &errMsg
		d.RetryCount++
	})
	w.notifyProgress()

	log.Errorf("Download %s failed: %v", w.download.ID, err)
	return err
}

// cancelCleanup nettoie après annulation
func (w *DownloadWorker) cancelCleanup() error {
	tempPath, _ := w.download.TempFilePath()
	os.Remove(tempPath)

	w.UpdateDownload(func(d *model.Download) {
		d.Status = model.StatusCancelled
	})
	w.notifyProgress()

	log.Infof("Download %s cancelled", w.download.ID)
	return nil
}

// cleanup nettoie les ressources
func (w *DownloadWorker) cleanup() {
	w.closeFile()

	if r := recover(); r != nil {
		log.Errorf("PANIC in download %s: %v", w.download.ID, r)
		w.fail(fmt.Errorf("panic: %v", r))
	}
}
