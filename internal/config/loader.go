package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

func (cfg *Config) setupLocalProviderOverrides() {
	for providerType, providerCfg := range cfg.AI.Providers {
		if providerCfg.MaxTokens == nil {
			providerCfg.MaxTokens = &cfg.AI.MaxTokens
		}

		if providerCfg.Temperature == nil {
			providerCfg.Temperature = &cfg.AI.Temperature
		}
		cfg.AI.Providers[providerType] = providerCfg
	}
}

// getConfigFile determines the base config file location using XDG standards.
func getConfigDir() (string, error) {
	configHome, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not find user config directory: %w", err)
	}
	configDir := filepath.Join(configHome, "commitgen")

	// Create the application's config directory if it doesn't exist.
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("could not create config directory at %s: %w", configDir, err)
	}

	configFile := filepath.Join(configDir, "config.toml")
	return configFile, nil
}

func GenerateConfig() error {
	configFile, err := getConfigDir()
	if err != nil {
		return err
	}

	// --- File doesn't exist: Write the default config ---
	data, err := toml.Marshal(NewDefaultConfig())
	if err != nil {
		return fmt.Errorf("could not marshal default config to TOML: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("could not write default config file: %w", err)
	}
	return nil
}

// LoadConfig finds, loads, and parses the configuration file.
func LoadConfig() (*Config, error) {
	configFile, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := GenerateConfig(); err != nil {
			return nil, err
		}
		cfg := NewDefaultConfig()
		cfg.setupLocalProviderOverrides()
		return cfg, nil
	}

	// --- File exists: Load it and merge with defaults ---
	cfg := NewDefaultConfig()
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not read config file at %s: %w", configFile, err)
	}

	/*
		Unmarshal the TOML data into the defaultConfig struct.
		Fields present in the file will override the default values.
	*/
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}
	cfg.setupLocalProviderOverrides()
	return cfg, nil
}
