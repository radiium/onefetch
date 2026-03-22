package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// ============================================================================
// FILESYSTEM UTILS
// ============================================================================

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
		return fmt.Errorf("file %s does not exist", src)
	}
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to rename file %s to %s: %w", src, dst, err)
	}

	return nil
}

// SamePath reports whether two paths (relative or absolute) point to the same file or directory.
func SamePath(path1, path2 string) (bool, error) {
	// Resolve both paths to absolute
	abs1, err := filepath.Abs(path1)
	if err != nil {
		return false, err
	}

	abs2, err := filepath.Abs(path2)
	if err != nil {
		return false, err
	}

	// Clean paths (resolve .., . and normalize separators)
	clean1 := filepath.Clean(abs1)
	clean2 := filepath.Clean(abs2)

	// Compare cleaned paths
	if clean1 == clean2 {
		return true, nil
	}

	// For a more robust check, use os.SameFile
	// which compares inodes on Unix and file IDs on Windows
	info1, err := os.Stat(clean1)
	if err != nil {
		// If the file does not exist, fall back to path comparison
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	info2, err := os.Stat(clean2)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// Use os.SameFile for physical inode comparison
	return os.SameFile(info1, info2), nil
}

// validatePathSafety performs common security checks on a path
func ValidatePathSafety(path string) (string, error) {
	// Get the absolute path and clean it
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check that it is not a symlink (if it exists)
	info, err := os.Lstat(absPath)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlinks are not allowed")
		}
	}

	return absPath, nil
}
