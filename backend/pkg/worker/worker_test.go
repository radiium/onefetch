package worker

import (
	"context"
	"dlbackend/internal/config"
	"dlbackend/internal/model"
	"dlbackend/pkg/client"
	"dlbackend/pkg/sse"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MOCK DOWNLOAD REPOSITORY
// ============================================================================

type MockDownloadRepository struct {
	mock.Mock
}

func (m *MockDownloadRepository) List(status []model.DownloadStatus, downloadTypes []model.DownloadType, page, limit int) ([]model.Download, int64, error) {
	args := m.Called(status, downloadTypes, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]model.Download), args.Get(1).(int64), args.Error(2)
}

func (m *MockDownloadRepository) Create(download *model.Download) error {
	args := m.Called(download)
	return args.Error(0)
}

func (m *MockDownloadRepository) GetByID(id string) (*model.Download, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Download), args.Error(1)
}

func (m *MockDownloadRepository) Update(download *model.Download) error {
	args := m.Called(download)
	return args.Error(0)
}

func (m *MockDownloadRepository) GetActive() ([]model.Download, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Download), args.Error(1)
}

func (m *MockDownloadRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// ============================================================================
// MOCK SETTINGS REPOSITORY
// ============================================================================

type MockSettingsRepository struct {
	mock.Mock
}

func (m *MockSettingsRepository) Get() (*model.Settings, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Settings), args.Error(1)
}

func (m *MockSettingsRepository) Update(settings *model.UpdateSettingsRequest) error {
	args := m.Called(settings)
	return args.Error(0)
}

// ============================================================================
// MOCK SSE MANAGER
// ============================================================================

type MockSSEManager struct {
	mock.Mock
}

func (m *MockSSEManager) GetClientCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSSEManager) GetClients() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

func (m *MockSSEManager) SendEvent(event string, data interface{}) error {
	args := m.Called(event, data)
	return args.Error(0)
}

func (m *MockSSEManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSSEManager) Print() {
	m.Called()
}

func (m *MockSSEManager) Handler(c *fiber.Ctx) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockSSEManager) FireHandlers(c *fiber.Ctx, event string) {
	m.Called(c, event)
}

func (m *MockSSEManager) OnConnect(handlers ...sse.OnStatusEventHandler) sse.Manager {
	args := m.Called(handlers)
	return args.Get(0).(sse.Manager)
}

func (m *MockSSEManager) OnDisconnect(handlers ...sse.OnStatusEventHandler) sse.Manager {
	args := m.Called(handlers)
	return args.Get(0).(sse.Manager)
}

func (m *MockSSEManager) OnEvent(eventName string, handlers ...sse.OnEventHandler) sse.Manager {
	args := m.Called(eventName, handlers)
	return args.Get(0).(sse.Manager)
}

// ============================================================================
// MOCK ONE FICHIER CLIENT
// ============================================================================

type MockOneFichierClient struct {
	mock.Mock
}

func (m *MockOneFichierClient) GetFileInfo(fileURL string) (*client.OneFichierInfoResponse, error) {
	args := m.Called(fileURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.OneFichierInfoResponse), args.Error(1)
}

func (m *MockOneFichierClient) GetDownloadToken(fileURL string) (*client.OneFichierTokenResponse, error) {
	args := m.Called(fileURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.OneFichierTokenResponse), args.Error(1)
}

func (m *MockOneFichierClient) DownloadFile(downloadURL string, offset int64) (io.ReadCloser, int64, int, error) {
	args := m.Called(downloadURL, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Get(2).(int), args.Error(3)
	}
	return args.Get(0).(io.ReadCloser), args.Get(1).(int64), args.Get(2).(int), args.Error(3)
}

// ============================================================================
// HELPER: Mock ReadCloser
// ============================================================================

type MockReadCloser struct {
	reader io.Reader
}

func (m *MockReadCloser) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *MockReadCloser) Close() error {
	return nil
}

// ============================================================================
// TEST HELPERS
// ============================================================================

func setupTestConfig(t *testing.T) string {
	tempDir := t.TempDir()
	config.Cfg = &config.Config{
		DLPath: tempDir,
	}
	return tempDir
}

// waitForWorkerCompletion attend que le worker soit supprimé de la map (fin d'exécution)
func waitForWorkerCompletion(t *testing.T, manager *DownloadManager, downloadID string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, exists := manager.workers.Load(downloadID)
			if !exists {
				// Worker terminé et supprimé
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("timeout waiting for worker %s to complete", downloadID)
			}
		}
	}
}

// ============================================================================
// DOWNLOAD MANAGER TESTS
// ============================================================================

func TestDownloadManager_Start(t *testing.T) {
	setupTestConfig(t)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockSettingsRepo := new(MockSettingsRepository)
		mockSSE := new(MockSSEManager)
		mockClient := new(MockOneFichierClient)

		mockSettingsRepo.On("Get").Return(&model.Settings{
			APIKey1fichier: "test-api-key",
		}, nil)

		// Mock pour stepGetFileInfo
		mockClient.On("GetFileInfo", "https://1fichier.com/test").Return(&client.OneFichierInfoResponse{
			Filename:    "test.pdf",
			Size:        int64(1024),
			Checksum:    "abc123",
			ContentType: "application/pdf",
		}, nil)

		// Mock pour stepGetDownloadToken
		mockClient.On("GetDownloadToken", "https://1fichier.com/test").Return(&client.OneFichierTokenResponse{
			URL: "https://download.1fichier.com/xyz",
		}, nil)

		// Mock repo.Update pour tous les appels
		mockRepo.On("Update", mock.MatchedBy(func(d *model.Download) bool {
			return d.ID == "test-id"
		})).Return(nil)

		// Mock SSE.SendEvent
		mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

		manager := NewDownloadManager(ctx, mockRepo, mockSettingsRepo, mockSSE)

		download := &model.Download{
			ID:      "test-id",
			FileURL: "https://1fichier.com/test",
			Status:  model.StatusPending,
			Type:    model.TypeMovie,
		}

		// Créer manuellement le worker avec le mock client au lieu de passer par manager.Start
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)
		manager.workers.Store(download.ID, worker)

		// Lancer le worker dans une goroutine (comme manager.Start le ferait)
		go func() {
			defer manager.workers.Delete(download.ID)
			worker.Run()
		}()

		// Attendre que le worker se termine
		waitForWorkerCompletion(t, manager, download.ID, 5*time.Second)

		// Vérifier que le worker a été supprimé après completion
		_, exists := manager.workers.Load(download.ID)
		assert.False(t, exists)

		// Vérifier que les mocks ont été appelés
		mockClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockSSE.AssertExpectations(t)
	})

	t.Run("no API key", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockSettingsRepo := new(MockSettingsRepository)
		mockSSE := new(MockSSEManager)

		mockSettingsRepo.On("Get").Return(&model.Settings{
			APIKey1fichier: "",
		}, nil)

		manager := NewDownloadManager(ctx, mockRepo, mockSettingsRepo, mockSSE)

		download := &model.Download{
			ID:      "test-id",
			FileURL: "https://1fichier.com/test",
			Type:    model.TypeMovie,
		}

		err := manager.Start(download)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key not configured")

		// Le worker ne doit pas être créé en cas d'erreur
		_, exists := manager.workers.Load(download.ID)
		assert.False(t, exists)
	})

	t.Run("settings error", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockSettingsRepo := new(MockSettingsRepository)
		mockSSE := new(MockSSEManager)

		mockSettingsRepo.On("Get").Return(nil, errors.New("db error"))

		manager := NewDownloadManager(ctx, mockRepo, mockSettingsRepo, mockSSE)

		download := &model.Download{
			ID:      "test-id",
			FileURL: "https://1fichier.com/test",
			Type:    model.TypeMovie,
		}

		err := manager.Start(download)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get settings")

		// Le worker ne doit pas être créé en cas d'erreur
		_, exists := manager.workers.Load(download.ID)
		assert.False(t, exists)
	})
}

func TestDownloadManager_Pause(t *testing.T) {
	setupTestConfig(t)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockClient := new(MockOneFichierClient)
		mockSSE := new(MockSSEManager)

		manager := &DownloadManager{
			ctx: ctx,
		}

		download := &model.Download{
			ID:   "test-id",
			Type: model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)
		manager.workers.Store(download.ID, worker)

		err := manager.Pause(download.ID)
		assert.NoError(t, err)
		assert.True(t, worker.IsPaused())
	})

	t.Run("download not found", func(t *testing.T) {
		ctx := context.Background()
		manager := &DownloadManager{ctx: ctx}

		err := manager.Pause("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "download not found")
	})
}

func TestDownloadManager_Resume(t *testing.T) {
	setupTestConfig(t)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockClient := new(MockOneFichierClient)
		mockSSE := new(MockSSEManager)

		manager := &DownloadManager{ctx: ctx}

		download := &model.Download{
			ID:   "test-id",
			Type: model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)
		worker.Pause()
		manager.workers.Store(download.ID, worker)

		err := manager.Resume(download.ID)
		assert.NoError(t, err)
		assert.False(t, worker.IsPaused())
	})

	t.Run("download not found", func(t *testing.T) {
		ctx := context.Background()
		manager := &DownloadManager{ctx: ctx}

		err := manager.Resume("non-existent")
		assert.Error(t, err)
	})
}

func TestDownloadManager_Cancel(t *testing.T) {
	setupTestConfig(t)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockClient := new(MockOneFichierClient)
		mockSSE := new(MockSSEManager)

		manager := &DownloadManager{ctx: ctx}

		download := &model.Download{
			ID:   "test-id",
			Type: model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)
		manager.workers.Store(download.ID, worker)

		err := manager.Cancel(download.ID)
		assert.NoError(t, err)
		assert.True(t, worker.IsCancelled())
	})

	t.Run("download not found", func(t *testing.T) {
		ctx := context.Background()
		manager := &DownloadManager{ctx: ctx}

		err := manager.Cancel("non-existent")
		assert.Error(t, err)
	})
}

// ============================================================================
// DOWNLOAD WORKER TESTS
// ============================================================================

func TestDownloadWorker_StateManagement(t *testing.T) {
	setupTestConfig(t)

	ctx := context.Background()
	mockRepo := new(MockDownloadRepository)
	mockClient := new(MockOneFichierClient)
	mockSSE := new(MockSSEManager)

	download := &model.Download{
		ID:   "test-id",
		Type: model.TypeMovie,
	}
	worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

	t.Run("initial state is running", func(t *testing.T) {
		assert.False(t, worker.IsPaused())
		assert.False(t, worker.IsCancelled())
	})

	t.Run("pause", func(t *testing.T) {
		worker.Pause()
		assert.True(t, worker.IsPaused())
		assert.False(t, worker.IsCancelled())
	})

	t.Run("resume", func(t *testing.T) {
		worker.Resume()
		assert.False(t, worker.IsPaused())
		assert.False(t, worker.IsCancelled())
	})

	t.Run("cancel", func(t *testing.T) {
		worker.Cancel()
		assert.True(t, worker.IsCancelled())
	})

	t.Run("cancel is idempotent", func(t *testing.T) {
		worker.Cancel()
		worker.Cancel()
		assert.True(t, worker.IsCancelled())
	})
}

func TestDownloadWorker_UpdateDownload(t *testing.T) {
	setupTestConfig(t)

	ctx := context.Background()
	mockRepo := new(MockDownloadRepository)
	mockClient := new(MockOneFichierClient)
	mockSSE := new(MockSSEManager)

	download := &model.Download{
		ID:              "test-id",
		DownloadedBytes: 0,
		Type:            model.TypeMovie,
	}
	worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

	// Test concurrent updates
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker.UpdateDownload(func(d *model.Download) {
				d.DownloadedBytes += 1
			})
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(100), download.DownloadedBytes)
}

func TestDownloadWorker_StepGetFileInfo(t *testing.T) {
	setupTestConfig(t)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockClient := new(MockOneFichierClient)
		mockSSE := new(MockSSEManager)

		fileSize := int64(1024)
		checksum := "abc123"
		contentType := "application/pdf"

		mockClient.On("GetFileInfo", "https://1fichier.com/test").Return(&client.OneFichierInfoResponse{
			Filename:    "test.pdf",
			Size:        fileSize,
			Checksum:    checksum,
			ContentType: contentType,
		}, nil)

		mockRepo.On("Update", mock.Anything).Return(nil)
		mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

		download := &model.Download{
			ID:      "test-id",
			FileURL: "https://1fichier.com/test",
			Type:    model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		err := worker.stepGetFileInfo()
		require.NoError(t, err)

		assert.Equal(t, "test.pdf", download.FileName)
		assert.Equal(t, fileSize, *download.FileSize)
		assert.Equal(t, checksum, *download.Checksum)
		assert.Equal(t, contentType, *download.MimeType)

		mockClient.AssertExpectations(t)
	})

	t.Run("client error", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockClient := new(MockOneFichierClient)
		mockSSE := new(MockSSEManager)

		mockClient.On("GetFileInfo", mock.Anything).Return(nil, errors.New("api error"))
		mockRepo.On("Update", mock.Anything).Return(nil)
		mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

		download := &model.Download{
			ID:      "test-id",
			FileURL: "https://1fichier.com/test",
			Type:    model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		err := worker.stepGetFileInfo()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get file info")
	})
}

func TestDownloadWorker_StepGetDownloadToken(t *testing.T) {
	setupTestConfig(t)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockClient := new(MockOneFichierClient)
		mockSSE := new(MockSSEManager)

		mockClient.On("GetDownloadToken", "https://1fichier.com/test").Return(&client.OneFichierTokenResponse{
			URL: "https://download.1fichier.com/xyz",
		}, nil)

		mockRepo.On("Update", mock.Anything).Return(nil)
		mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

		download := &model.Download{
			ID:      "test-id",
			FileURL: "https://1fichier.com/test",
			Type:    model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		err := worker.stepGetDownloadToken()
		require.NoError(t, err)

		assert.NotNil(t, download.DownloadURL)
		assert.Equal(t, "https://download.1fichier.com/xyz", *download.DownloadURL)
		assert.NotNil(t, download.DownloadURLExpiresAt)

		mockClient.AssertExpectations(t)
	})

	t.Run("client error", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockClient := new(MockOneFichierClient)
		mockSSE := new(MockSSEManager)

		mockClient.On("GetDownloadToken", mock.Anything).Return(nil, errors.New("token error"))
		mockRepo.On("Update", mock.Anything).Return(nil)
		mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

		download := &model.Download{
			ID:      "test-id",
			FileURL: "https://1fichier.com/test",
			Type:    model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		err := worker.stepGetDownloadToken()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get download token")
	})
}

func TestDownloadWorker_CalculateTotalSize(t *testing.T) {
	setupTestConfig(t)

	ctx := context.Background()
	mockRepo := new(MockDownloadRepository)
	mockClient := new(MockOneFichierClient)
	mockSSE := new(MockSSEManager)

	t.Run("with existing file size", func(t *testing.T) {
		fileSize := int64(1000)
		download := &model.Download{
			ID:       "test-id",
			FileSize: &fileSize,
			Type:     model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		size := worker.calculateTotalSize(http.StatusOK, 500)
		assert.Equal(t, int64(1000), size)
	})

	t.Run("partial content", func(t *testing.T) {
		download := &model.Download{
			ID:              "test-id",
			DownloadedBytes: 500,
			Type:            model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		size := worker.calculateTotalSize(http.StatusPartialContent, 500)
		assert.Equal(t, int64(1000), size)
		assert.Equal(t, int64(1000), *download.FileSize)
	})

	t.Run("full content resets offset", func(t *testing.T) {
		download := &model.Download{
			ID:              "test-id",
			DownloadedBytes: 500,
			Type:            model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		size := worker.calculateTotalSize(http.StatusOK, 1000)
		assert.Equal(t, int64(1000), size)
		assert.Equal(t, int64(0), download.DownloadedBytes)
	})
}

func TestDownloadWorker_PrepareFile(t *testing.T) {
	setupTestConfig(t)

	ctx := context.Background()
	mockRepo := new(MockDownloadRepository)
	mockClient := new(MockOneFichierClient)
	mockSSE := new(MockSSEManager)

	t.Run("create new file", func(t *testing.T) {
		download := &model.Download{
			ID:              "test-id",
			DownloadedBytes: 0,
			FileName:        "test.txt",
			Type:            model.TypeMovie,
		}
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		err := worker.prepareFile()
		require.NoError(t, err)
		assert.NotNil(t, worker.file)

		worker.closeFile()
	})

	t.Run("resume existing file", func(t *testing.T) {
		download := &model.Download{
			ID:              "test-id",
			DownloadedBytes: 0,
			FileName:        "test.txt",
			Type:            model.TypeMovie,
		}

		// Créer un fichier existant
		tempPath, _ := download.TempFilePath()
		os.MkdirAll(filepath.Dir(tempPath), 0755)
		existingData := []byte("existing data")
		os.WriteFile(tempPath, existingData, 0644)

		download.DownloadedBytes = int64(len(existingData))
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		err := worker.prepareFile()
		require.NoError(t, err)
		assert.NotNil(t, worker.file)

		worker.closeFile()
	})

	t.Run("resume with custom directory", func(t *testing.T) {
		customDir := "custom_folder"
		download := &model.Download{
			ID:              "test-id",
			DownloadedBytes: 0,
			FileName:        "test.txt",
			CustomFileDir:   &customDir,
			Type:            model.TypeMovie,
		}

		tempPath, _ := download.TempFilePath()
		os.MkdirAll(filepath.Dir(tempPath), 0755)
		os.WriteFile(tempPath, []byte("data"), 0644)

		download.DownloadedBytes = 4
		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		err := worker.prepareFile()
		require.NoError(t, err)

		worker.closeFile()
	})
}

func TestDownloadWorker_DownloadChunk(t *testing.T) {
	setupTestConfig(t)

	t.Run("successful download", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockClient := new(MockOneFichierClient)
		mockSSE := new(MockSSEManager)

		downloadURL := "https://download.1fichier.com/test"
		fileSize := int64(100)

		download := &model.Download{
			ID:              "test-id",
			DownloadURL:     &downloadURL,
			FileName:        "test.txt",
			FileSize:        &fileSize,
			DownloadedBytes: 0,
			Type:            model.TypeMovie,
		}

		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

		// Préparer le fichier
		err := worker.prepareFile()
		require.NoError(t, err)
		defer worker.closeFile()

		// Mock le client
		testData := []byte("test data content")
		reader := &MockReadCloser{reader: strings.NewReader(string(testData))}
		mockClient.On("DownloadFile", downloadURL, int64(0)).Return(
			reader,
			int64(len(testData)),
			http.StatusOK,
			nil,
		)

		// Mock Update pour chaque appel
		mockRepo.On("Update", mock.MatchedBy(func(d *model.Download) bool {
			return d.ID == download.ID
		})).Return(nil)

		// Mock SendEvent pour chaque appel
		mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

		completed, err := worker.downloadChunk()
		require.NoError(t, err)
		assert.True(t, completed)
		assert.Equal(t, int64(len(testData)), download.DownloadedBytes)

		mockClient.AssertExpectations(t)
	})

	t.Run("cancelled during download", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := new(MockDownloadRepository)
		mockClient := new(MockOneFichierClient)
		mockSSE := new(MockSSEManager)

		downloadURL := "https://download.1fichier.com/test"

		download := &model.Download{
			ID:          "test-id",
			DownloadURL: &downloadURL,
			FileName:    "test.txt",
			Type:        model.TypeMovie,
		}

		worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)
		worker.Cancel()

		err := worker.prepareFile()
		require.NoError(t, err)
		defer worker.closeFile()

		reader := &MockReadCloser{reader: strings.NewReader("test")}
		mockClient.On("DownloadFile", downloadURL, int64(0)).Return(
			reader, int64(4), http.StatusOK, nil,
		)

		completed, err := worker.downloadChunk()
		assert.Error(t, err)
		assert.False(t, completed)
		assert.Contains(t, err.Error(), "cancelled")
	})
}

func TestDownloadWorker_UpdateSpeed(t *testing.T) {
	setupTestConfig(t)

	ctx := context.Background()
	mockRepo := new(MockDownloadRepository)
	mockClient := new(MockOneFichierClient)
	mockSSE := new(MockSSEManager)

	mockRepo.On("Update", mock.Anything).Return(nil)
	mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

	download := &model.Download{
		ID:              "test-id",
		DownloadedBytes: 1000,
		Type:            model.TypeMovie,
	}
	worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

	lastUpdate := time.Now().Add(-1 * time.Second)
	lastBytes := int64(0)

	worker.updateSpeed(&lastUpdate, &lastBytes)

	assert.NotNil(t, download.Speed)
	assert.Greater(t, *download.Speed, float64(0))
}

func TestDownloadWorker_Complete(t *testing.T) {
	setupTestConfig(t)

	ctx := context.Background()
	mockRepo := new(MockDownloadRepository)
	mockClient := new(MockOneFichierClient)
	mockSSE := new(MockSSEManager)

	mockRepo.On("Update", mock.Anything).Return(nil)
	mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

	download := &model.Download{
		ID:       "test-id",
		FileName: "test.txt",
		Status:   model.StatusDownloading,
		Type:     model.TypeMovie,
	}

	// Créer le fichier temp
	tempPath, _ := download.TempFilePath()
	os.MkdirAll(filepath.Dir(tempPath), 0755)
	os.WriteFile(tempPath, []byte("test content"), 0644)

	worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

	err := worker.complete()
	require.NoError(t, err)

	assert.Equal(t, model.StatusCompleted, download.Status)
	assert.Equal(t, float64(100), download.Progress)
	assert.NotNil(t, download.CompletedAt)

	// Vérifier que le fichier final existe
	finalPath, _ := download.FinalFilePath()
	_, err = os.Stat(finalPath)
	assert.NoError(t, err)

	// Vérifier que le fichier temp n'existe plus
	_, err = os.Stat(tempPath)
	assert.True(t, os.IsNotExist(err))
}

func TestDownloadWorker_Fail(t *testing.T) {
	setupTestConfig(t)

	ctx := context.Background()
	mockRepo := new(MockDownloadRepository)
	mockClient := new(MockOneFichierClient)
	mockSSE := new(MockSSEManager)

	mockRepo.On("Update", mock.Anything).Return(nil)
	mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

	download := &model.Download{
		ID:         "test-id",
		Status:     model.StatusDownloading,
		RetryCount: 0,
		Type:       model.TypeMovie,
	}

	worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

	testErr := errors.New("test error")
	err := worker.fail(testErr)

	assert.Error(t, err)
	assert.Equal(t, model.StatusFailed, download.Status)
	assert.NotNil(t, download.ErrorMessage)
	assert.Equal(t, "test error", *download.ErrorMessage)
	assert.Equal(t, 1, download.RetryCount)
}

func TestDownloadWorker_CancelCleanup(t *testing.T) {
	setupTestConfig(t)

	ctx := context.Background()
	mockRepo := new(MockDownloadRepository)
	mockClient := new(MockOneFichierClient)
	mockSSE := new(MockSSEManager)

	mockRepo.On("Update", mock.Anything).Return(nil)
	mockSSE.On("SendEvent", "progress", mock.Anything).Return(nil)

	download := &model.Download{
		ID:       "test-id",
		FileName: "test.txt",
		Status:   model.StatusDownloading,
		Type:     model.TypeMovie,
	}

	// Créer le fichier temp
	tempPath, _ := download.TempFilePath()
	os.MkdirAll(filepath.Dir(tempPath), 0755)
	os.WriteFile(tempPath, []byte("test"), 0644)

	worker := NewDownloadWorker(ctx, download, mockRepo, mockClient, mockSSE)

	err := worker.cancelCleanup()
	require.NoError(t, err)

	assert.Equal(t, model.StatusCancelled, download.Status)

	// Vérifier que le fichier temp est supprimé
	_, err = os.Stat(tempPath)
	assert.True(t, os.IsNotExist(err))
}
