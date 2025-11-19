package utils

import (
	"dlbackend/internal/model"
	"fmt"
	"net/url"
	"strings"
)

// ============================================================================
// VALIDATION UTILS
// ============================================================================

// Validate1FichierURL trim and validate 1fichier.com URL
//   - cannot be empty
//   - must be a valid URL
//   - must begin with https scheme
//   - must be a valide domain name
//   - must contains query param
func Validate1FichierURL(rawURL string) (string, error) {
	urlStr := strings.TrimSpace(rawURL)

	// Basic validation - the URL must not be empty
	if urlStr == "" {
		return "", fmt.Errorf("URL is required")
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %s", urlStr)
	}

	// Check URL scheme
	if parsedURL.Scheme != "https" {
		return "", fmt.Errorf("invalid URL scheme: %s", parsedURL.Scheme)
	}

	// Check the exact domain name
	host := strings.ToLower(parsedURL.Host)
	if host != "1fichier.com" && host != "www.1fichier.com" {
		return "", fmt.Errorf("unauthorized domain: %s", parsedURL.Host)
	}

	// Checks for the presence of query param (1fichier.com file ID)
	if parsedURL.RawQuery == "" {
		return "", fmt.Errorf("1fichier id not found in URL: %s", rawURL)
	}

	return urlStr, nil
}

// ValidateType convert string input to DownloadType and validate
func ValidateType(typeStr string) (model.DownloadType, error) {
	typeStr = strings.TrimSpace(typeStr)
	dt := model.DownloadType(typeStr)
	switch dt {
	case model.TypeMovie, model.TypeSerie:
		return dt, nil
	default:
		return "", fmt.Errorf("invalid type: %s", typeStr)
	}
}

// ValidateNotEmpty trim the string value and check if it's empty
func ValidateNotEmpty(name string, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("'%s' is required", name)
	}
	return value, nil
}

// ValidatePath trim and validates path format
//   - cannot be empty
//   - cannot contains more then 4096 characters
//   - cannot contains null bytes characters
func ValidatePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	// check the length
	if len(path) > 4096 {
		return "", fmt.Errorf("path too long (max 4096 characters)")
	}

	// basic format checks
	if strings.Contains(path, "\x00") {
		return "", fmt.Errorf("path contains null bytes")
	}

	return path, nil
}

const (
	maxPathLength    = 4096 // PATH_MAX on Linux
	maxSegmentLength = 255  // NAME_MAX on Linux
	maxDepth         = 10   // Maximum folder depth
)

// ValidateFolderName validates and normalizes a user-provided folder name.
// Returns the normalized path (trailing slash removed) or an error.
// Empty strings are accepted and returned as-is.
//
// Validation rules:
//   - Accepts empty strings
//   - Allows subdirectories via "/" separator
//   - Rejects absolute paths (starting with "/")
//   - Rejects parent directory references ("..")
//   - Rejects hidden folders (starting with ".")
//   - Rejects control characters and null bytes
//   - Enforces maximum path length, segment length, and depth
func ValidateDirName(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	original := path
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimSpace(path)

	if path == "" {
		return "", nil
	}

	if len(path) > maxPathLength {
		return "", fmt.Errorf("path exceeds maximum length of %d bytes: '%s'", maxPathLength, original)
	}

	if strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("absolute paths are not allowed: '%s'", original)
	}

	segments := strings.Split(path, "/")

	if len(segments) > maxDepth {
		return "", fmt.Errorf("path exceeds maximum depth of %d levels: '%s'", maxDepth, original)
	}

	for _, segment := range segments {
		if err := validateSegment(segment, original); err != nil {
			return "", err
		}
	}

	return path, nil
}

// validateSegment checks if a single path segment is valid.
func validateSegment(segment, originalPath string) error {
	if segment == "" {
		return fmt.Errorf("empty segments are not allowed: '%s'", originalPath)
	}

	if strings.TrimSpace(segment) == "" {
		return fmt.Errorf("segments containing only whitespace are not allowed: '%s'", originalPath)
	}

	if len(segment) > maxSegmentLength {
		return fmt.Errorf("segment exceeds maximum length of %d bytes: '%s'", maxSegmentLength, originalPath)
	}

	if segment == ".." {
		return fmt.Errorf("parent directory references (..) are not allowed: '%s'", originalPath)
	}

	if strings.HasPrefix(segment, ".") {
		return fmt.Errorf("hidden folders are not allowed: '%s'", originalPath)
	}

	for _, c := range segment {
		if c == 0 || c < 32 || c == 127 {
			return fmt.Errorf("control characters are not allowed in path: '%s'", originalPath)
		}
	}

	return nil
}

// ValidateFileName validates a user-provided file name.
// Returns the normalized name (trimmed) or an error.
// Empty strings are accepted and returned as-is.
//
// Validation rules:
//   - Accepts empty strings
//   - Rejects path separators ("/" and "\")
//   - Rejects parent directory references ("..")
//   - Rejects current directory reference (".")
//   - Rejects hidden files (starting with ".")
//   - Rejects control characters and null bytes
//   - Enforces maximum name length
func ValidateFileName(name string) (string, error) {
	if name == "" {
		return "", nil
	}

	original := name
	name = strings.TrimSpace(name)

	if name == "" {
		return "", nil
	}

	const maxNameLength = 255 // Standard filesystem limit

	if len(name) > maxNameLength {
		return "", fmt.Errorf("file name exceeds maximum length of %d bytes: '%s'", maxNameLength, original)
	}

	// Reject path separators
	if strings.ContainsAny(name, "/\\") {
		return "", fmt.Errorf("file name cannot contain path separators: '%s'", original)
	}

	// Reject parent/current directory references
	if name == ".." || name == "." {
		return "", fmt.Errorf("file name cannot be '.' or '..': '%s'", original)
	}

	// Reject hidden files (starting with .)
	if strings.HasPrefix(name, ".") {
		return "", fmt.Errorf("hidden files (starting with '.') are not allowed: '%s'", original)
	}

	// Reject control characters and null bytes
	for _, r := range name {
		if r < 32 || r == 127 || r == 0 {
			return "", fmt.Errorf("file name contains invalid control characters: '%s'", original)
		}
	}

	return name, nil
}
