package config

import (
	"os"
	"path/filepath"
)

const (
	// AppName defines the name of the application.
	AppName = "sirrmeshd"
	// Version defines the version of the application.
	Version = "0.3.1"
)

// MustGetDefaultNodeHome returns the default node home directory.
// It defaults to $HOME/.sirrmeshd or uses MAILCHAT_HOME environment variable if set.
func MustGetDefaultNodeHome() string {
	// Check for environment variable first
	if envHome := os.Getenv("MAILCHAT_HOME"); envHome != "" {
		return envHome
	}

	// Get user home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir cannot be determined
		return ".sirrmeshd"
	}

	return filepath.Join(userHome, ".sirrmeshd")
}
