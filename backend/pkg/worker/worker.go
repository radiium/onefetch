package worker

import (
	"context"
	"dlbackend/internal/config"
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

	"github.com/gofiber/fiber/v3/log"
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
		return fmt.Errorf("API key not configured")
	}

	client := client.NewOneFichierClient(config.Cfg.ApiUrl1fichier, settings.APIKey1fichier)
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

	// Safe: workers only stores *DownloadWorker values (see Start).
	worker := value.(*DownloadWorker)
	worker.Pause()
	return nil
}

func (m *DownloadManager) Resume(downloadID string) error {
	value, ok := m.workers.Load(downloadID)
	if !ok {
		return errors.New("download not found")
	}

	// Safe: workers only stores *DownloadWorker values (see Start).
	worker := value.(*DownloadWorker)
	worker.Resume()
	return nil
}

func (m *DownloadManager) Cancel(downloadID string) error {
	value, ok := m.workers.Load(downloadID)
	if !ok {
		return errors.New("download not found")
	}

	// Safe: workers only stores *DownloadWorker values (see Start).
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

	// State control via atomics (no mutex needed)
	state atomic.Int32 // 0=running, 1=paused, 2=cancelled

	ctx    context.Context
	cancel context.CancelFunc

	// File writing
	file *os.File
	mu   sync.Mutex // Only for file operations
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

// Pause requests a pause (thread-safe, non-blocking).
func (w *DownloadWorker) Pause() {
	w.state.Store(StatePaused)
	log.Infof("Pause requested for download %s", w.download.ID)
}

// Resume cancels the pause (thread-safe, non-blocking).
func (w *DownloadWorker) Resume() {
	if w.state.CompareAndSwap(StatePaused, StateRunning) {
		log.Infof("Resume requested for download %s", w.download.ID)
	}
}

// Cancel stops the download (idempotent).
func (w *DownloadWorker) Cancel() {
	if w.state.Swap(StateCancelled) != StateCancelled {
		log.Infof("Cancel requested for download %s", w.download.ID)
		w.cancel()
	}
}

// IsPaused reports whether the worker is currently paused.
func (w *DownloadWorker) IsPaused() bool {
	return w.state.Load() == StatePaused
}

// IsCancelled reports whether the worker has been cancelled.
func (w *DownloadWorker) IsCancelled() bool {
	return w.state.Load() == StateCancelled
}

// UpdateDownload applies fn to the download struct (thread-safe).
func (w *DownloadWorker) UpdateDownload(fn func(*model.Download)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fn(w.download)
}

// notifyProgress persists the current download state to the DB and broadcasts an SSE progress event.
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
		DownloadedBytes: w.download.DownloadedBytes,
		FileSize:        w.download.FileSize,
		Speed:           w.download.Speed,
	}

	if err := w.sseManager.SendEvent("progress", event); err != nil {
		log.Errorf("Failed to send SSE for download %s: %v", w.download.ID, err)
	}
}

// Run executes the full download workflow sequentially.
func (w *DownloadWorker) Run() error {
	defer w.cleanup()

	// Sequential steps
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

// stepGetFileInfo fetches file metadata from the 1fichier API.
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

// stepGetDownloadToken fetches a time-limited download token from the 1fichier API.
// WARNING: the token is valid for 5 minutes only. Downloads longer than 5 minutes
// will fail mid-transfer. Token renewal on expiry is not yet implemented.
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

// stepDownload performs the actual file download with pause/resume support.
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

	// Open or create the temp file
	if err := w.prepareFile(); err != nil {
		return err
	}
	defer w.closeFile()

	// Download loop: retries automatically after each pause
	for {
		if w.IsCancelled() {
			return errors.New("cancelled")
		}

		// Wait here while paused
		if w.IsPaused() {
			w.UpdateDownload(func(d *model.Download) {
				d.Status = model.StatusPaused
			})
			w.notifyProgress()

			log.Debugf("Download %s paused, waiting for resume...", w.download.ID)

			// Active wait with periodic state check.
			// NOTE: intentional busy-wait polling every 100ms.
			// Consider replacing with sync.Cond.Wait() if CPU usage becomes a concern.
			for w.IsPaused() && !w.IsCancelled() {
				time.Sleep(100 * time.Millisecond)
			}

			if w.IsCancelled() {
				return errors.New("cancelled")
			}

			// Resume
			w.UpdateDownload(func(d *model.Download) {
				d.Status = model.StatusDownloading
			})
			w.notifyProgress()
			log.Debugf("Download %s resumed", w.download.ID)
		}

		// Download a chunk
		completed, err := w.downloadChunk()
		if err != nil {
			return err
		}

		if completed {
			return nil
		}
	}
}

// prepareFile opens or creates the temp file for writing, resuming from the current offset if applicable.
func (w *DownloadWorker) prepareFile() error {
	tempPath, err := w.download.TempFilePath()
	if err != nil {
		return fmt.Errorf("failed to resolve temp path: %w", err)
	}

	// Create the directory
	if err := os.MkdirAll(filepath.Dir(tempPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if the temp file exists and matches the expected offset (resume support)
	if w.download.DownloadedBytes > 0 {
		stat, err := os.Stat(tempPath)
		if err != nil || stat.Size() != w.download.DownloadedBytes {
			log.Warnf("Temp file mismatch, restarting from zero: %s", w.download.ID)
			w.UpdateDownload(func(d *model.Download) {
				d.DownloadedBytes = 0
			})
		}
	}

	// Open in append mode for resume, or create a new file
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

// downloadChunk downloads data from the current offset until EOF, pause, or cancel.
func (w *DownloadWorker) downloadChunk() (completed bool, err error) {
	reader, contentLength, statusCode, err := w.client.DownloadFile(
		*w.download.DownloadURL,
		w.download.DownloadedBytes,
	)
	if err != nil {
		return false, fmt.Errorf("failed to start download: %w", err)
	}
	defer reader.Close()

	// Calculate total size from metadata or response headers
	totalSize := w.calculateTotalSize(statusCode, contentLength)

	// Read and write with periodic state checks
	buffer := make([]byte, 64*1024)
	lastUpdate := time.Now()
	lastBytes := w.download.DownloadedBytes
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Check state
		if w.IsCancelled() {
			return false, errors.New("cancelled")
		}

		if w.IsPaused() {
			log.Debugf("Pause detected during download chunk for %s", w.download.ID)
			return false, nil // Return without error; the outer loop handles the pause
		}

		// Update speed periodically
		select {
		case <-ticker.C:
			w.updateSpeed(&lastUpdate, &lastBytes)
		default:
		}

		// Lire
		n, readErr := reader.Read(buffer)

		if n > 0 {
			// Write
			if _, err := w.file.Write(buffer[:n]); err != nil {
				return false, fmt.Errorf("failed to write: %w", err)
			}

			// Update progress
			w.UpdateDownload(func(d *model.Download) {
				d.DownloadedBytes += int64(n)
				if totalSize > 0 {
					d.Progress = float64(d.DownloadedBytes) / float64(totalSize) * 100
				}
			})
		}

		if readErr == io.EOF {
			w.updateSpeed(&lastUpdate, &lastBytes)
			return true, nil // Download complete
		}

		if readErr != nil {
			return false, readErr
		}
	}
}

// calculateTotalSize resolves the total file size from known metadata or response headers.
// If the server returns 200 instead of 206 (Partial Content), it does not support Range
// requests; the offset is reset to zero and the download restarts from the beginning.
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

// updateSpeed recalculates download speed (bytes/sec) since the last call.
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

// closeFile flushes and closes the temp file.
func (w *DownloadWorker) closeFile() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		w.file.Sync()
		w.file.Close()
		w.file = nil
	}
}

// complete renames the temp file to its final path and marks the download as completed.
func (w *DownloadWorker) complete() error {
	tempPath, _ := w.download.TempFilePath()
	finalPath, _ := w.download.FinalFilePath()

	// Remove the final file if it already exists (overwrite)
	os.Remove(finalPath)

	// Rename temp to final path
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

// fail marks the download as failed and broadcasts the error via SSE.
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

// cancelCleanup removes the temp file and marks the download as cancelled.
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

// cleanup closes open files and recovers from panics.
func (w *DownloadWorker) cleanup() {
	w.closeFile()

	if r := recover(); r != nil {
		log.Errorf("PANIC in download %s: %v", w.download.ID, r)
		w.fail(fmt.Errorf("panic: %v", r))
	}
}
