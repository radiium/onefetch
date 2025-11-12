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

// SamePath vérifie si deux chemins (relatifs ou absolus) pointent vers le même fichier/dossier
func SamePath(path1, path2 string) (bool, error) {
	// Convertir les deux chemins en chemins absolus
	abs1, err := filepath.Abs(path1)
	if err != nil {
		return false, err
	}

	abs2, err := filepath.Abs(path2)
	if err != nil {
		return false, err
	}

	// Nettoyer les chemins (résoudre .. et . et normaliser les séparateurs)
	clean1 := filepath.Clean(abs1)
	clean2 := filepath.Clean(abs2)

	// Comparer les chemins nettoyés
	if clean1 == clean2 {
		return true, nil
	}

	// Pour une vérification plus robuste, on peut aussi utiliser os.SameFile
	// qui compare les inodes sur Unix et les file IDs sur Windows
	info1, err := os.Stat(clean1)
	if err != nil {
		// Si le fichier n'existe pas, on se fie à la comparaison des chemins
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

	// Utiliser os.SameFile pour la comparaison physique
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
