package utils

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

func WriteTempFileWithContext(ctx context.Context, src io.ReadCloser, tempPath string, progressCallback func(int64)) error {
	defer src.Close()

	if err := EnsureDir(filepath.Dir(tempPath)); err != nil {
		return err
	}

	file, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return copyWithContext(ctx, file, src, progressCallback)
}

func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader, callback func(int64)) error {
	buf := make([]byte, 32*1024) // 32KB chunks
	var totalBytes int64

	for {
		// Check cancel context
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
