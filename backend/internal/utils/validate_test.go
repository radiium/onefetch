package utils

import (
	"dlbackend/internal/model"
	"testing"
)

func TestValidate1FichierURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid URL with www",
			input:   "https://www.1fichier.com/?abc123",
			want:    "https://www.1fichier.com/?abc123",
			wantErr: false,
		},
		{
			name:    "valid URL without www",
			input:   "https://1fichier.com/?xyz789",
			want:    "https://1fichier.com/?xyz789",
			wantErr: false,
		},
		{
			name:    "valid URL with whitespace",
			input:   "  https://1fichier.com/?test  ",
			want:    "https://1fichier.com/?test",
			wantErr: false,
		},
		{
			name:    "empty URL",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			input:   "not a url",
			wantErr: true,
		},
		{
			name:    "http scheme instead of https",
			input:   "http://1fichier.com/?abc",
			wantErr: true,
		},
		{
			name:    "wrong domain",
			input:   "https://example.com/?abc",
			wantErr: true,
		},
		{
			name:    "missing query param",
			input:   "https://1fichier.com/",
			wantErr: true,
		},
		{
			name:    "subdomain not allowed",
			input:   "https://api.1fichier.com/?abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Validate1FichierURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate1FichierURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Validate1FichierURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    model.DownloadType
		wantErr bool
	}{
		{
			name:    "valid movie type",
			input:   "MOVIE",
			want:    model.TypeMovie,
			wantErr: false,
		},
		{
			name:    "valid serie type",
			input:   "SERIE",
			want:    model.TypeSerie,
			wantErr: false,
		},
		{
			name:    "movie with whitespace",
			input:   "  MOVIE  ",
			want:    model.TypeMovie,
			wantErr: false,
		},
		{
			name:    "invalid type",
			input:   "documentary",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "wrong case",
			input:   "Movie",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ValidateType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateNotEmpty(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		input     string
		want      string
		wantErr   bool
	}{
		{
			name:      "valid non-empty string",
			fieldName: "username",
			input:     "john",
			want:      "john",
			wantErr:   false,
		},
		{
			name:      "string with whitespace trimmed",
			fieldName: "title",
			input:     "  test  ",
			want:      "test",
			wantErr:   false,
		},
		{
			name:      "empty string",
			fieldName: "email",
			input:     "",
			wantErr:   true,
		},
		{
			name:      "whitespace only",
			fieldName: "description",
			input:     "   ",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateNotEmpty(tt.fieldName, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotEmpty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ValidateNotEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid path",
			input:   "/home/user/documents",
			want:    "/home/user/documents",
			wantErr: false,
		},
		{
			name:    "path with whitespace trimmed",
			input:   "  /var/log  ",
			want:    "/var/log",
			wantErr: false,
		},
		{
			name:    "empty path",
			input:   "",
			wantErr: true,
		},
		{
			name:    "path with null byte",
			input:   "/home/user\x00/file",
			wantErr: true,
		},
		{
			name:    "path too long",
			input:   string(make([]byte, 4097)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidatePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ValidatePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateDirName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "empty string allowed",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "simple directory name",
			input:   "videos",
			want:    "videos",
			wantErr: false,
		},
		{
			name:    "nested path",
			input:   "movies/2024",
			want:    "movies/2024",
			wantErr: false,
		},
		{
			name:    "trailing slash removed",
			input:   "documents/",
			want:    "documents",
			wantErr: false,
		},
		{
			name:    "whitespace trimmed",
			input:   "  folder  ",
			want:    "folder",
			wantErr: false,
		},
		{
			name:    "absolute path rejected",
			input:   "/home/user",
			wantErr: true,
		},
		{
			name:    "parent directory reference rejected",
			input:   "../other",
			wantErr: true,
		},
		{
			name:    "hidden folder rejected",
			input:   ".hidden",
			wantErr: true,
		},
		{
			name:    "hidden nested folder rejected",
			input:   "folder/.hidden",
			wantErr: true,
		},
		{
			name:    "control characters rejected",
			input:   "folder\x00name",
			wantErr: true,
		},
		{
			name:    "too deep path",
			input:   "a/b/c/d/e/f/g/h/i/j/k",
			wantErr: true,
		},
		{
			name:    "segment too long",
			input:   string(make([]byte, 256)),
			wantErr: true,
		},
		{
			name:    "empty segment rejected",
			input:   "folder//subfolder",
			wantErr: true,
		},
		{
			name:    "whitespace only segment rejected",
			input:   "folder/ /subfolder",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateDirName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ValidateDirName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "empty string allowed",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "valid filename",
			input:   "document.txt",
			want:    "document.txt",
			wantErr: false,
		},
		{
			name:    "filename with whitespace trimmed",
			input:   "  file.pdf  ",
			want:    "file.pdf",
			wantErr: false,
		},
		{
			name:    "path separator forward slash rejected",
			input:   "folder/file.txt",
			wantErr: true,
		},
		{
			name:    "path separator backslash rejected",
			input:   "folder\\file.txt",
			wantErr: true,
		},
		{
			name:    "parent directory reference rejected",
			input:   "..",
			wantErr: true,
		},
		{
			name:    "current directory reference rejected",
			input:   ".",
			wantErr: true,
		},
		{
			name:    "hidden file rejected",
			input:   ".hidden",
			wantErr: true,
		},
		{
			name:    "control character rejected",
			input:   "file\x00.txt",
			wantErr: true,
		},
		{
			name:    "filename too long",
			input:   string(make([]byte, 256)),
			wantErr: true,
		},
		{
			name:    "whitespace only becomes empty and allowed",
			input:   "   ",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateFileName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ValidateFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateSegment(t *testing.T) {
	tests := []struct {
		name         string
		segment      string
		originalPath string
		wantErr      bool
	}{
		{
			name:         "valid segment",
			segment:      "folder",
			originalPath: "folder",
			wantErr:      false,
		},
		{
			name:         "empty segment",
			segment:      "",
			originalPath: "folder//subfolder",
			wantErr:      true,
		},
		{
			name:         "whitespace only segment",
			segment:      "   ",
			originalPath: "folder/   /subfolder",
			wantErr:      true,
		},
		{
			name:         "segment too long",
			segment:      string(make([]byte, 256)),
			originalPath: "path",
			wantErr:      true,
		},
		{
			name:         "parent directory reference",
			segment:      "..",
			originalPath: "../folder",
			wantErr:      true,
		},
		{
			name:         "hidden folder",
			segment:      ".hidden",
			originalPath: ".hidden",
			wantErr:      true,
		},
		{
			name:         "control character",
			segment:      "folder\x00",
			originalPath: "folder\x00",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSegment(tt.segment, tt.originalPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSegment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
