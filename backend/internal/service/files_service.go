package service

import (
	"dlbackend/internal/config"
	"dlbackend/internal/errors"
	"dlbackend/internal/model"
	"dlbackend/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2/log"
)

type FilesService interface {
	GetDir() (*model.FSNode, error)
	CreateDir(path string, dirName string) (*model.FSNode, error)
	DeleteDir(path string) (*model.FSNode, error)
}

type filesService struct {
}

func NewFilesService() FilesService {
	return &filesService{}
}

func (fs *filesService) GetDir() (*model.FSNode, error) {
	return fs.getDirTree()
}

func (fs *filesService) CreateDir(path string, dirName string) (*model.FSNode, error) {
	// Validate parent path
	absPath, err := fs.buildAbsPath(path)
	if err != nil {
		log.Error(err)
		return nil, errors.BadRequest(fmt.Sprintf("invalid parent path: %s", path))
	}

	// Build new directory path
	newDirPath := filepath.Join(absPath, dirName)

	// Check if directory already exists
	if _, err := os.Stat(newDirPath); err == nil {
		return nil, errors.Conflict(fmt.Sprintf("directory '%s' already exists", dirName))
	}

	// Create directory
	if err := utils.EnsureDir(newDirPath); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to create directory '%s': %v", dirName, err))
	}

	// Verify created path is safe (defense in depth)
	if _, err := fs.buildAbsPath(newDirPath); err != nil {
		// Rollback: delete the created directory
		os.RemoveAll(newDirPath)
		return nil, errors.Forbidden("security check failed after directory creation")
	}

	return fs.getDirTree()
}

func (fs *filesService) DeleteDir(path string) (*model.FSNode, error) {
	// Validate and obtain the absolute path (includes symlink check)
	absPath, err := fs.buildAbsPath(path)
	if err != nil {
		log.Error(err)
		return nil, errors.BadRequest(fmt.Sprintf("invalid path: %s", path))
	}

	// Verify that the path exists
	if _, err := os.Lstat(absPath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NotFound(fmt.Sprintf("directory not found: %s", path))
		}
		return nil, errors.Internal(fmt.Sprintf("cannot access path '%s': %v", path, err))
	}

	// Remove directory
	if err := os.RemoveAll(absPath); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to delete directory '%s': %v", path, err))
	}

	return fs.getDirTree()
}

// Helpers

// getDirTree retrieves the complete directory tree and handles errors
func (fs *filesService) getDirTree() (*model.FSNode, error) {
	tree, err := fs.buildDirTree(config.Cfg.DLPath)
	if err != nil {
		log.Error(err)
		return nil, errors.Internal(fmt.Sprintf("failed to get directory tree: %v", err))
	}
	return &tree, nil
}

// buildDirTree generates a tree representation of the file system for a given path
func (vs *filesService) buildDirTree(root string) (model.FSNode, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return model.FSNode{}, err
	}

	// Reject symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		return model.FSNode{}, fmt.Errorf("symlinks are not allowed")
	}

	node := model.FSNode{
		Name:  info.Name(),
		Path:  root,
		IsDir: info.IsDir(),
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
			childNode, err := vs.buildDirTree(childPath)
			if err == nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node, nil
}

// buildAbsPath generate the absolute path
// and verify that it is contained within one of the authorized folders.
func (fs *filesService) buildAbsPath(path string) (string, error) {
	// Authorized directories
	allowedDirs := []string{
		filepath.Join(config.Cfg.DLPath, model.TypeMovie.Dir()),
		filepath.Join(config.Cfg.DLPath, model.TypeSerie.Dir()),
	}

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

	// Verify that the path is a subdirectory of an allowed directory
	for _, dir := range allowedDirs {
		absAllowed, err := filepath.Abs(dir)
		if err != nil {
			continue
		}

		// Ensure both paths end with separator for accurate prefix matching
		allowedPrefix := absAllowed + string(os.PathSeparator)
		if strings.HasPrefix(absPath+string(os.PathSeparator), allowedPrefix) && absPath != absAllowed {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("unauthorized: path outside allowed directories")
}
