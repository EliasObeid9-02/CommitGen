package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

/*
LoadConfig finds, loads, and parses the configuration file.

It follows the XDG Base Directory Specification to locate the config file.
*/
func LoadConfig() (*Config, error) {
	// Determine the base config directory using XDG standards.
	configHome, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not find user config directory: %w", err)
	}
	configDir := filepath.Join(configHome, "commitgen")

	// Create the application's config directory if it doesn't exist.
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create config directory at %s: %w", configDir, err)
	}

	configFile := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// --- File doesn't exist: Create and write the default config ---
		defaultConfig := NewDefaultConfig()

		data, err := toml.Marshal(defaultConfig)
		if err != nil {
			return nil, fmt.Errorf("could not marshal default config to TOML: %w", err)
		}

		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return nil, fmt.Errorf("could not write default config file: %w", err)
		}
		return defaultConfig, nil
	}

	// --- File exists: Load it and merge with defaults ---
	defaultConfig := NewDefaultConfig()
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not read config file at %s: %w", configFile, err)
	}

	/*
		Unmarshal the TOML data into the defaultConfig struct.
		Fields present in the file will override the default values.
	*/
	if err := toml.Unmarshal(data, defaultConfig); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}
	return defaultConfig, nil
}
