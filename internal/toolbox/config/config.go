package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Session stores the current user session
type Session struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// GetConfigDir returns the config directory path
func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".nf-tools", "config")
}

// GetBinDir returns the bin directory path
func GetBinDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".nf-tools", "bin")
}

// GetCacheDir returns the cache directory path
func GetCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".nf-tools", "cache")
}

// EnsureDirs creates all required directories
func EnsureDirs() error {
	dirs := []string{GetConfigDir(), GetBinDir(), GetCacheDir()}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}

// LoadSession reads the session from disk
func LoadSession() (*Session, error) {
	path := filepath.Join(GetConfigDir(), "session.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// SaveSession writes the session to disk
func SaveSession(session *Session) error {
	if err := EnsureDirs(); err != nil {
		return err
	}

	path := filepath.Join(GetConfigDir(), "session.json")

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// DeleteSession removes the session file
func DeleteSession() error {
	path := filepath.Join(GetConfigDir(), "session.json")
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Settings stores user preferences
type Settings struct {
	AutoUpdate     bool   `json:"auto_update"`
	CheckUpdates   bool   `json:"check_updates"`
	TelemetryOptIn bool   `json:"telemetry_opt_in"`
	APIBaseURL     string `json:"api_base_url,omitempty"` // For dev/testing
}

// DefaultSettings returns default settings
func DefaultSettings() *Settings {
	return &Settings{
		AutoUpdate:     false,
		CheckUpdates:   true,
		TelemetryOptIn: false,
	}
}

// LoadSettings reads settings from disk
func LoadSettings() (*Settings, error) {
	path := filepath.Join(GetConfigDir(), "settings.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultSettings(), nil
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return DefaultSettings(), nil
	}

	return &settings, nil
}

// SaveSettings writes settings to disk
func SaveSettings(settings *Settings) error {
	if err := EnsureDirs(); err != nil {
		return err
	}

	path := filepath.Join(GetConfigDir(), "settings.json")

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
