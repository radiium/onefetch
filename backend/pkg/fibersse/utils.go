package fibersse

import (
	"fmt"
	"regexp"
)

// ValidateChannelName valide le nom d'un canal
func ValidateChannelBasePath(name string) error {
	if name == "" {
		return fmt.Errorf("channel name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("channel name cannot exceed 100 characters")
	}
	return nil
}

// ValidateBasePath valide le chemin de base du canal
func ValidateBasePath(base string) error {
	if base == "" {
		return fmt.Errorf("base path cannot be empty")
	}
	if base[0] != '/' {
		return fmt.Errorf("base path must start with /")
	}
	validPath := regexp.MustCompile(`^/[a-zA-Z0-9/_-]*$`)
	if !validPath.MatchString(base) {
		return fmt.Errorf("base path contains invalid characters")
	}
	return nil
}
