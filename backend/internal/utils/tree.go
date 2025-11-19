package utils

import (
	"dlbackend/internal/config"
	"dlbackend/internal/model"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================================
// FILESYSTEM TREE UTILS
// ============================================================================

// buildDirTree generates a tree representation of the file system for a given path
func BuildDirTree(root string) (model.FSNode, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return model.FSNode{}, err
	}

	// Reject symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		return model.FSNode{}, fmt.Errorf("symlinks are not allowed")
	}

	// Check read only directories
	isRootDir, err := SamePath(info.Name(), config.Cfg.DLPath)
	if err != nil {
		return model.FSNode{}, err
	}
	isMovieDir, err := SamePath(info.Name(), filepath.Join(config.Cfg.DLPath, model.TypeMovie.Dir()))
	if err != nil {
		return model.FSNode{}, err
	}
	isSerieDir, err := SamePath(info.Name(), filepath.Join(config.Cfg.DLPath, model.TypeSerie.Dir()))
	if err != nil {
		return model.FSNode{}, err
	}

	node := model.FSNode{
		Name:       info.Name(),
		Path:       root,
		IsDir:      info.IsDir(),
		IsHidden:   isHiddenFile(info.Name()),
		IsTmp:      isTmpFile(info.Name()),
		IsReadOnly: isRootDir || isMovieDir || isSerieDir,
	}

	if info.IsDir() {
		entries, err := os.ReadDir(root)
		if err != nil {
			return node, err
		}

		for _, entry := range entries {
			// Ignore symlinks in the directory tree
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}

			childPath := filepath.Join(root, entry.Name())
			childNode, err := BuildDirTree(childPath)
			if err == nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node, nil
}

func isHiddenFile(path string) bool {
	name := filepath.Base(path)
	return strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".tmp")
}
func isTmpFile(path string) bool {
	name := filepath.Base(path)
	return isHiddenFile(name) && strings.HasSuffix(name, ".tmp")
}

// BuildDirList Returns a list of all folders and subfolders
// within a given path, traversed recursively.
func BuildDirTreeAsList(rootPath string) ([]string, error) {
	var directories []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If it's a folder and not the root folder
		if info.IsDir() && path != rootPath {
			// Get the path relative to the root path
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}
			directories = append(directories, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return directories, nil
}
