package config

import (
	"os"
	"path/filepath"
	"testing"
)

/*
TestFindEditor uses a table-driven approach to test the findEditor function
by simulating different environment variable states.
*/
func TestFindEditor(t *testing.T) {
	testCases := []struct {
		name     string
		visual   string
		editor   string
		expected string
	}{
		{
			name:     "VISUAL is set",
			visual:   "vim",
			editor:   "nano",
			expected: "vim",
		},
		{
			name:     "VISUAL is unset EDITOR is set",
			visual:   "",
			editor:   "emacs",
			expected: "emacs",
		},
		{
			name:     "Both are unset",
			visual:   "",
			editor:   "",
			expected: "nano",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("VISUAL", tc.visual)
			os.Setenv("EDITOR", tc.editor)

			if result := findEditor(); result != tc.expected {
				t.Errorf("expected %q, but got %q", tc.expected, result)
			}
		})
	}
}

/*
TestLoadConfig covers the primary scenarios for the configuration loader.
It uses a temporary directory to avoid interfering with the actual user config.
*/
func TestLoadConfig(t *testing.T) {
	t.Run("First run with no existing config", func(t *testing.T) {
		// Create a temporary directory for the test and override XDG_CONFIG_HOME
		tempDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tempDir)

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		// Check that the returned config is the default one
		if cfg.DefaultType != "refactor" {
			t.Errorf("expected default type 'refactor', got %q", cfg.DefaultType)
		}

		// Verify the config file was actually created
		expectedPath := filepath.Join(tempDir, "commitgen", "config.toml")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected config file to be created at %s, but it wasn't", expectedPath)
		}
	})

	t.Run("Existing config overrides default values", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tempDir)

		// Manually create a config file with custom values
		configDir := filepath.Join(tempDir, "commitgen")
		os.MkdirAll(configDir, 0755)
		configFile := filepath.Join(configDir, "config.toml")

		customConfigContent := `
default_type = "feat"
commit_username = "Test User"

[ai]
  max_tokens = 999
`
		if err := os.WriteFile(configFile, []byte(customConfigContent), 0644); err != nil {
			t.Fatalf("failed to write custom config file: %v", err)
		}

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		// Check that custom values correctly override defaults
		if cfg.DefaultType != "feat" {
			t.Errorf("expected default_type to be 'feat', got %q", cfg.DefaultType)
		}
		if cfg.CommitUserName != "Test User" {
			t.Errorf("expected commit_username to be 'Test User', got %q", cfg.CommitUserName)
		}
		if cfg.AI.MaxTokens != 999 {
			t.Errorf("expected ai.max_tokens to be 999, got %d", cfg.AI.MaxTokens)
		}

		// Check that non-overridden values remain at their defaults
		if cfg.AI.Temperature != 0.3 {
			t.Errorf("expected ai.temperature to be default 0.3, got %f", cfg.AI.Temperature)
		}
	})

	t.Run("Malformed config file returns an error", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tempDir)

		// Create a malformed config file
		configDir := filepath.Join(tempDir, "commitgen")
		os.MkdirAll(configDir, 0755)
		configFile := filepath.Join(configDir, "config.toml")

		malformedContent := `default_type = "oops this is not valid toml`
		if err := os.WriteFile(configFile, []byte(malformedContent), 0644); err != nil {
			t.Fatalf("failed to write malformed config file: %v", err)
		}

		_, err := LoadConfig()
		if err == nil {
			t.Fatal("expected an error for malformed config file, but got nil")
		}
	})
}
