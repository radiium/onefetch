package filesystem

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type FileManager interface {
	EnsureDir(path string) error
	WriteTempFileWithContext(ctx context.Context, src io.ReadCloser, tempPath string, progressCallback func(int64)) error
	MoveFile(src, dst string) error
	RemoveFile(path string) error
}

type fileManager struct{}

func NewFileManager() FileManager {
	return &fileManager{}
}

func (fm *fileManager) EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func (fm *fileManager) WriteTempFileWithContext(ctx context.Context, src io.ReadCloser, tempPath string, progressCallback func(int64)) error {
	defer src.Close()

	if err := fm.EnsureDir(filepath.Dir(tempPath)); err != nil {
		return err
	}

	file, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return fm.copyWithContext(ctx, file, src, progressCallback)
}

func (fm *fileManager) MoveFile(src, dst string) error {
	if err := fm.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

func (fm *fileManager) RemoveFile(path string) error {
	return os.Remove(path)
}

func (fm *fileManager) copyWithContext(ctx context.Context, dst io.Writer, src io.Reader, callback func(int64)) error {
	buf := make([]byte, 32*1024) // 32KB chunks
	var totalBytes int64

	for {
		// Vérifier le contexte d'annulation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			totalBytes += int64(n)
			if callback != nil {
				callback(totalBytes)
			}
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
