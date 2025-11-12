package utils

import (
	"dlbackend/internal/config"
	"dlbackend/internal/model"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"
)

func TestBuildDirTree(t *testing.T) {
	config.Load()

	// Créer un répertoire temporaire pour les tests
	tmpDir, err := os.MkdirTemp("", "test_build_dir_tree")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup: créer une structure de répertoires et fichiers
	// tmpDir/
	//   ├── file1.txt
	//   ├── subdir1/
	//   │   ├── file2.txt
	//   │   └── subdir2/
	//   │       └── file3.txt
	//   └── emptydir/

	// Créer les fichiers et répertoires
	file1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	subdir1 := filepath.Join(tmpDir, "subdir1")
	if err := os.Mkdir(subdir1, 0755); err != nil {
		t.Fatalf("Failed to create subdir1: %v", err)
	}

	file2 := filepath.Join(subdir1, "file2.txt")
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	subdir2 := filepath.Join(subdir1, "subdir2")
	if err := os.Mkdir(subdir2, 0755); err != nil {
		t.Fatalf("Failed to create subdir2: %v", err)
	}

	file3 := filepath.Join(subdir2, "file3.txt")
	if err := os.WriteFile(file3, []byte("content3"), 0644); err != nil {
		t.Fatalf("Failed to create file3: %v", err)
	}

	emptydir := filepath.Join(tmpDir, "emptydir")
	if err := os.Mkdir(emptydir, 0755); err != nil {
		t.Fatalf("Failed to create emptydir: %v", err)
	}

	tests := []struct {
		name           string
		setup          func() string
		wantErr        bool
		errContains    string
		validateResult func(*testing.T, model.FSNode)
	}{
		{
			name:    "single file",
			setup:   func() string { return file1 },
			wantErr: false,
			validateResult: func(t *testing.T, node model.FSNode) {
				if node.Name != "file1.txt" {
					t.Errorf("expected Name 'file1.txt', got '%s'", node.Name)
				}
				if node.IsDir {
					t.Errorf("expected IsDir false, got true")
				}
				if len(node.Children) != 0 {
					t.Errorf("expected no children, got %d", len(node.Children))
				}
			},
		},
		{
			name:    "directory with files",
			setup:   func() string { return tmpDir },
			wantErr: false,
			validateResult: func(t *testing.T, node model.FSNode) {
				if !node.IsDir {
					t.Errorf("expected IsDir true, got false")
				}
				if len(node.Children) != 3 {
					t.Errorf("expected 3 children, got %d", len(node.Children))
				}
				// Vérifier que les enfants sont présents
				foundFile1 := false
				foundSubdir1 := false
				foundEmptydir := false
				for _, child := range node.Children {
					switch child.Name {
					case "file1.txt":
						foundFile1 = true
						if child.IsDir {
							t.Errorf("file1.txt should not be a directory")
						}
					case "subdir1":
						foundSubdir1 = true
						if !child.IsDir {
							t.Errorf("subdir1 should be a directory")
						}
					case "emptydir":
						foundEmptydir = true
						if !child.IsDir {
							t.Errorf("emptydir should be a directory")
						}
						if len(child.Children) != 0 {
							t.Errorf("emptydir should have no children")
						}
					}
				}
				if !foundFile1 || !foundSubdir1 || !foundEmptydir {
					t.Errorf("missing expected children")
				}
			},
		},
		{
			name:    "nested directories",
			setup:   func() string { return subdir1 },
			wantErr: false,
			validateResult: func(t *testing.T, node model.FSNode) {
				if node.Name != "subdir1" {
					t.Errorf("expected Name 'subdir1', got '%s'", node.Name)
				}
				if !node.IsDir {
					t.Errorf("expected IsDir true, got false")
				}
				if len(node.Children) != 2 {
					t.Errorf("expected 2 children, got %d", len(node.Children))
				}
			},
		},
		{
			name:    "empty directory",
			setup:   func() string { return emptydir },
			wantErr: false,
			validateResult: func(t *testing.T, node model.FSNode) {
				if !node.IsDir {
					t.Errorf("expected IsDir true, got false")
				}
				if len(node.Children) != 0 {
					t.Errorf("expected 0 children, got %d", len(node.Children))
				}
			},
		},
		{
			name:        "non-existent path",
			setup:       func() string { return filepath.Join(tmpDir, "nonexistent") },
			wantErr:     true,
			errContains: "",
		},
		{
			name: "symlink is rejected",
			setup: func() string {
				if runtime.GOOS == "windows" {
					t.Skip("Skipping symlink test on Windows")
				}
				symlinkPath := filepath.Join(tmpDir, "symlink.txt")
				if err := os.Symlink(file1, symlinkPath); err != nil {
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

			result, err := BuildDirTree(path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("error = %v, should contain %v", err, tt.errContains)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error = %v", err)
				}
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
				// Vérifications communes
				if result.Path != path {
					t.Errorf("expected Path '%s', got '%s'", path, result.Path)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestBuildDirTreeAsList(t *testing.T) {
	// Créer un répertoire temporaire pour les tests
	tmpDir, err := os.MkdirTemp("", "get_all_dirs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		setup    func(string) error
		want     []string
		wantErr  bool
		rootPath string
	}{
		{
			name: "dossier vide",
			setup: func(root string) error {
				return nil
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "un seul sous-dossier",
			setup: func(root string) error {
				return os.Mkdir(filepath.Join(root, "subdir"), 0755)
			},
			want:    []string{"subdir"},
			wantErr: false,
		},
		{
			name: "plusieurs sous-dossiers au même niveau",
			setup: func(root string) error {
				dirs := []string{"dir1", "dir2", "dir3"}
				for _, dir := range dirs {
					if err := os.Mkdir(filepath.Join(root, dir), 0755); err != nil {
						return err
					}
				}
				return nil
			},
			want:    []string{"dir1", "dir2", "dir3"},
			wantErr: false,
		},
		{
			name: "dossiers imbriqués",
			setup: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "a", "b", "c"), 0755); err != nil {
					return err
				}
				return nil
			},
			want:    []string{"a", filepath.Join("a", "b"), filepath.Join("a", "b", "c")},
			wantErr: false,
		},
		{
			name: "structure complexe avec fichiers",
			setup: func(root string) error {
				// Créer des dossiers
				dirs := []string{
					"dir1",
					filepath.Join("dir1", "subdir1"),
					filepath.Join("dir1", "subdir2"),
					"dir2",
					filepath.Join("dir2", "subdir3"),
				}
				for _, dir := range dirs {
					if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
						return err
					}
				}

				// Créer des fichiers (ne doivent pas être listés)
				files := []string{
					"file.txt",
					filepath.Join("dir1", "file1.txt"),
					filepath.Join("dir1", "subdir1", "file2.txt"),
				}
				for _, file := range files {
					if err := os.WriteFile(filepath.Join(root, file), []byte("test"), 0644); err != nil {
						return err
					}
				}

				return nil
			},
			want: []string{
				"dir1",
				filepath.Join("dir1", "subdir1"),
				filepath.Join("dir1", "subdir2"),
				"dir2",
				filepath.Join("dir2", "subdir3"),
			},
			wantErr: false,
		},
		{
			name: "dossiers avec noms spéciaux",
			setup: func(root string) error {
				dirs := []string{
					".hidden",
					"dir with spaces",
					"dir-with-dash",
					"dir_with_underscore",
				}
				for _, dir := range dirs {
					if err := os.Mkdir(filepath.Join(root, dir), 0755); err != nil {
						return err
					}
				}
				return nil
			},
			want: []string{
				".hidden",
				"dir with spaces",
				"dir-with-dash",
				"dir_with_underscore",
			},
			wantErr: false,
		},
		{
			name:     "chemin inexistant",
			setup:    func(root string) error { return nil },
			want:     nil,
			wantErr:  true,
			rootPath: "/path/that/does/not/exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Créer un sous-dossier temporaire pour chaque test
			testDir := tmpDir
			if tt.rootPath == "" {
				testDir = filepath.Join(tmpDir, tt.name)
				if err := os.MkdirAll(testDir, 0755); err != nil {
					t.Fatal(err)
				}
			} else {
				testDir = tt.rootPath
			}

			// Configurer la structure de test
			if tt.setup != nil {
				if err := tt.setup(testDir); err != nil {
					t.Fatal(err)
				}
			}

			// Exécuter la fonction
			got, err := BuildDirTreeAsList(testDir)

			// Vérifier l'erreur
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildDirTreeAsList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Trier pour la comparaison
			sort.Strings(got)
			sort.Strings(tt.want)

			// Normaliser les slices vides (nil vs []string{})
			if len(got) == 0 && len(tt.want) == 0 {
				return // Les deux sont vides, c'est OK
			}

			// Comparer les résultats
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildDirTreeAsList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildDirTreeAsListPermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir, err := os.MkdirTemp("", "get_all_dirs_perm_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Créer un dossier sans permissions de lecture
	noReadDir := filepath.Join(tmpDir, "noread")
	if err := os.Mkdir(noReadDir, 0755); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(noReadDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Retirer les permissions de lecture
	if err := os.Chmod(noReadDir, 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(noReadDir, 0755) // Restaurer pour le nettoyage

	_, err = BuildDirTreeAsList(tmpDir)
	if err == nil {
		t.Error("BuildDirTreeAsList() expected error for directory without read permissions")
	}
}
