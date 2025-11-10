package utils

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mockReadCloser simule un io.ReadCloser pour les tests
type mockReadCloser struct {
	data   []byte
	offset int
	closed bool
}

func newMockReadCloser(data string) *mockReadCloser {
	return &mockReadCloser{
		data:   []byte(data),
		offset: 0,
		closed: false,
	}
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.offset >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.offset:])
	m.offset += n
	return n, nil
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return nil
}

// slowReadCloser simule une lecture lente pour tester l'annulation
type slowReadCloser struct {
	data   []byte
	offset int
	delay  time.Duration
	closed bool
}

func newSlowReadCloser(data string, delay time.Duration) *slowReadCloser {
	return &slowReadCloser{
		data:   []byte(data),
		offset: 0,
		delay:  delay,
		closed: false,
	}
}

func (s *slowReadCloser) Read(p []byte) (n int, err error) {
	time.Sleep(s.delay)
	if s.offset >= len(s.data) {
		return 0, io.EOF
	}
	n = copy(p, s.data[s.offset:])
	s.offset += n
	return n, nil
}

func (s *slowReadCloser) Close() error {
	s.closed = true
	return nil
}

// errorReadCloser simule une erreur de lecture
type errorReadCloser struct {
	err error
}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorReadCloser) Close() error {
	return nil
}

func TestWriteTempFileWithContext_Success(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "subdir", "test.txt")
	testData := "Hello, World! This is test data."

	src := newMockReadCloser(testData)
	var progressUpdates []int64

	progressCallback := func(bytes int64) {
		progressUpdates = append(progressUpdates, bytes)
	}

	ctx := context.Background()
	err := WriteTempFileWithContext(ctx, src, tempPath, progressCallback)

	if err != nil {
		t.Fatalf("WriteTempFileWithContext failed: %v", err)
	}

	if !src.closed {
		t.Error("Source reader was not closed")
	}

	content, err := os.ReadFile(tempPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("File content mismatch. Expected: %s, Got: %s", testData, string(content))
	}

	if len(progressUpdates) == 0 {
		t.Error("Progress callback was not called")
	}

	lastProgress := progressUpdates[len(progressUpdates)-1]
	if lastProgress != int64(len(testData)) {
		t.Errorf("Final progress mismatch. Expected: %d, Got: %d", len(testData), lastProgress)
	}
}

func TestWriteTempFileWithContext_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "test.txt")

	// Créer un fichier assez large et avec un délai suffisant pour garantir l'annulation
	largeData := strings.Repeat("A", 5*1024*1024) // 5MB
	src := newSlowReadCloser(largeData, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	// Annuler le contexte après un court délai
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := WriteTempFileWithContext(ctx, src, tempPath, nil)

	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestWriteTempFileWithContext_ReadError(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "test.txt")

	expectedErr := errors.New("read error")
	src := &errorReadCloser{err: expectedErr}

	ctx := context.Background()
	err := WriteTempFileWithContext(ctx, src, tempPath, nil)

	if err == nil {
		t.Fatal("Expected read error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error: %v, Got: %v", expectedErr, err)
	}
}

func TestWriteTempFileWithContext_InvalidPath(t *testing.T) {
	// Utiliser un chemin invalide (en supposant que /root/noperm n'est pas accessible)
	tempPath := "/root/noperm/test.txt"
	src := newMockReadCloser("test data")

	ctx := context.Background()
	err := WriteTempFileWithContext(ctx, src, tempPath, nil)

	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}
}

func TestWriteTempFileWithContext_NilProgressCallback(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "test.txt")
	testData := "Test without callback"

	src := newMockReadCloser(testData)
	ctx := context.Background()

	err := WriteTempFileWithContext(ctx, src, tempPath, nil)

	if err != nil {
		t.Fatalf("WriteTempFileWithContext failed: %v", err)
	}

	content, err := os.ReadFile(tempPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("File content mismatch. Expected: %s, Got: %s", testData, string(content))
	}
}

func TestWriteTempFileWithContext_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "empty.txt")

	src := newMockReadCloser("")
	ctx := context.Background()

	err := WriteTempFileWithContext(ctx, src, tempPath, nil)

	if err != nil {
		t.Fatalf("WriteTempFileWithContext failed: %v", err)
	}

	info, err := os.Stat(tempPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != 0 {
		t.Errorf("Expected empty file, got size: %d", info.Size())
	}
}

func TestWriteTempFileWithContext_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "large.txt")

	// Créer un fichier de 1MB
	largeData := strings.Repeat("X", 1024*1024)
	src := newMockReadCloser(largeData)

	var progressCount int
	progressCallback := func(bytes int64) {
		progressCount++
	}

	ctx := context.Background()
	err := WriteTempFileWithContext(ctx, src, tempPath, progressCallback)

	if err != nil {
		t.Fatalf("WriteTempFileWithContext failed: %v", err)
	}

	info, err := os.Stat(tempPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != int64(len(largeData)) {
		t.Errorf("File size mismatch. Expected: %d, Got: %d", len(largeData), info.Size())
	}

	if progressCount == 0 {
		t.Error("Progress callback was not called for large file")
	}
}

func TestCopyWithContext_ProgressTracking(t *testing.T) {
	testData := "This is progress tracking test data"
	src := strings.NewReader(testData)

	var dst strings.Builder
	var progressUpdates []int64

	progressCallback := func(bytes int64) {
		progressUpdates = append(progressUpdates, bytes)
	}

	ctx := context.Background()
	err := copyWithContext(ctx, &dst, src, progressCallback)

	if err != nil {
		t.Fatalf("copyWithContext failed: %v", err)
	}

	if dst.String() != testData {
		t.Errorf("Data mismatch. Expected: %s, Got: %s", testData, dst.String())
	}

	if len(progressUpdates) == 0 {
		t.Error("Progress callback was not called")
	}

	// Vérifier que les progrès sont cumulatifs
	for i := 1; i < len(progressUpdates); i++ {
		if progressUpdates[i] <= progressUpdates[i-1] {
			t.Errorf("Progress not cumulative at index %d: %v", i, progressUpdates)
		}
	}
}
