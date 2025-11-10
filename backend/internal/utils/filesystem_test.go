package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDir(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create simple directory",
			path:    filepath.Join(t.TempDir(), "test_dir"),
			wantErr: false,
		},
		{
			name:    "create nested directories",
			path:    filepath.Join(t.TempDir(), "parent", "child", "grandchild"),
			wantErr: false,
		},
		{
			name:    "directory already exists",
			path:    t.TempDir(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureDir(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if _, err := os.Stat(tt.path); os.IsNotExist(err) {
					t.Errorf("EnsureDir() directory was not created: %s", tt.path)
				}
			}
		})
	}
}

func TestMoveFile(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (src, dst string)
		wantErr bool
	}{
		{
			name: "move file successfully",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				src := filepath.Join(tmpDir, "source.txt")
				dst := filepath.Join(tmpDir, "destination.txt")

				if err := os.WriteFile(src, []byte("test content"), 0644); err != nil {
					t.Fatal(err)
				}

				return src, dst
			},
			wantErr: false,
		},
		{
			name: "move file to nested directory",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				src := filepath.Join(tmpDir, "source.txt")
				dst := filepath.Join(tmpDir, "subdir", "nested", "destination.txt")

				if err := os.WriteFile(src, []byte("test content"), 0644); err != nil {
					t.Fatal(err)
				}

				return src, dst
			},
			wantErr: false,
		},
		{
			name: "source file does not exist",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				src := filepath.Join(tmpDir, "nonexistent.txt")
				dst := filepath.Join(tmpDir, "destination.txt")
				return src, dst
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setup(t)

			err := MoveFile(src, dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoveFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify source no longer exists
				if _, err := os.Stat(src); !os.IsNotExist(err) {
					t.Errorf("MoveFile() source file still exists: %s", src)
				}

				// Verify destination exists
				if _, err := os.Stat(dst); os.IsNotExist(err) {
					t.Errorf("MoveFile() destination file does not exist: %s", dst)
				}
			}
		})
	}
}

func TestGetDirectories(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		wantDirs []string
		wantErr  bool
	}{
		{
			name: "get directories from path with multiple folders",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Create directories
				os.Mkdir(filepath.Join(tmpDir, "dir1"), 0755)
				os.Mkdir(filepath.Join(tmpDir, "dir2"), 0755)
				os.Mkdir(filepath.Join(tmpDir, "dir3"), 0755)

				// Create files (should be ignored)
				os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
				os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content"), 0644)

				return tmpDir
			},
			wantDirs: []string{"dir1", "dir2", "dir3"},
			wantErr:  false,
		},
		{
			name: "empty directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantDirs: []string{},
			wantErr:  false,
		},
		{
			name: "directory with only files",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
				os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content"), 0644)
				return tmpDir
			},
			wantDirs: []string{},
			wantErr:  false,
		},
		{
			name: "nonexistent directory",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantDirs: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)

			got, err := GetDirectories(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDirectories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.wantDirs) {
					t.Errorf("GetDirectories() got %d directories, want %d", len(got), len(tt.wantDirs))
					return
				}

				// Convert to map for easier comparison
				gotMap := make(map[string]bool)
				for _, dir := range got {
					gotMap[dir] = true
				}

				for _, wantDir := range tt.wantDirs {
					if !gotMap[wantDir] {
						t.Errorf("GetDirectories() missing directory: %s", wantDir)
					}
				}
			}
		})
	}
}
