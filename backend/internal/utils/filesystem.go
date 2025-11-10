package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir create a directory at the specified location if it does not exist.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to initialize directory %s: %w", path, err)
	}
	return nil
}

// MoveFile move and ensure destination directory exists
func MoveFile(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		fmt.Println("Le fichier n'existe pas")
		return fmt.Errorf("file %s does not exist", src)
	}
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to rename file %s to %s: %w:", src, dst, err)
	}

	return nil
}

// GetDirectories returns a list of folders in the specified path
func GetDirectories(path string) ([]string, error) {
	var directories []string

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	return directories, nil
}
