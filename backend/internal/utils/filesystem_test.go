package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestSamePath(t *testing.T) {
	// Créer un répertoire temporaire pour les tests
	tmpDir, err := os.MkdirTemp("", "samepath_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Créer des fichiers et dossiers de test
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Créer un lien symbolique
	linkFile := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(testFile, linkFile); err != nil {
		t.Logf("Symlink creation skipped: %v", err)
	}

	tests := []struct {
		name    string
		path1   string
		path2   string
		want    bool
		wantErr bool
	}{
		{
			name:  "chemins identiques absolus",
			path1: testFile,
			path2: testFile,
			want:  true,
		},
		{
			name:  "même fichier avec ./",
			path1: testFile,
			path2: "./" + testFile,
			want:  false,
		},
		{
			name:  "même fichier avec ../",
			path1: testFile,
			path2: filepath.Join(tmpDir, "..", filepath.Base(tmpDir), "test.txt"),
			want:  true,
		},
		{
			name:  "fichiers différents",
			path1: testFile,
			path2: filepath.Join(tmpDir, "other.txt"),
			want:  false,
		},
		{
			name:  "dossiers identiques",
			path1: testDir,
			path2: testDir,
			want:  true,
		},
		{
			name:  "fichier vs dossier",
			path1: testFile,
			path2: testDir,
			want:  false,
		},
		{
			name:  "lien symbolique vers même fichier",
			path1: testFile,
			path2: linkFile,
			want:  true,
		},
		{
			name:  "chemin avec /.",
			path1: testFile,
			path2: testFile + "/.",
			want:  true,
		},
		{
			name:    "fichier inexistant",
			path1:   filepath.Join(tmpDir, "nonexistent1.txt"),
			path2:   filepath.Join(tmpDir, "nonexistent2.txt"),
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SamePath(tt.path1, tt.path2)
			if (err != nil) != tt.wantErr {
				t.Errorf("SamePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SamePath() = %v, want %v", got, tt.want)
				t.Logf("path1: %s", tt.path1)
				t.Logf("path2: %s", tt.path2)
			}
		})
	}
}

func TestValidatePathSafety(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_validate_path")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		setup       func() string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid absolute path",
			setup:   func() string { return testFile },
			wantErr: false,
		},
		{
			name: "valid relative path",
			setup: func() string {
				rel, _ := filepath.Rel(".", testFile)
				return rel
			},
			wantErr: false,
		},
		{
			name:    "path with dots gets cleaned",
			setup:   func() string { return filepath.Join(tmpDir, ".", "test.txt") },
			wantErr: false,
		},
		{
			name:    "path with double dots gets cleaned",
			setup:   func() string { return filepath.Join(tmpDir, "subdir", "..", "test.txt") },
			wantErr: false,
		},
		{
			name:    "non-existent path is allowed",
			setup:   func() string { return filepath.Join(tmpDir, "nonexistent.txt") },
			wantErr: false,
		},
		{
			name: "symlink is rejected",
			setup: func() string {
				if runtime.GOOS == "windows" {
					t.Skip("Skipping symlink test on Windows")
				}
				symlinkPath := filepath.Join(tmpDir, "symlink.txt")
				if err := os.Symlink(testFile, symlinkPath); err != nil {
					t.Fatalf("Failed to create symlink: %v", err)
				}
				return symlinkPath
			},
			wantErr:     true,
			errContains: "symlinks are not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result, err := ValidatePathSafety(path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if tt.errContains != "" && err != nil && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, should contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error = %v", err)
				}
				if result == "" {
					t.Errorf("returned empty path")
				}
				if !filepath.IsAbs(result) {
					t.Errorf("result is not absolute: %v", result)
				}
			}
		})
	}
}
