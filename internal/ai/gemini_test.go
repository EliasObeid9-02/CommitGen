package ai

import (
	"CommitGen/internal/config"
	"context"
	"os"
	"strings"
	"testing"
)

const stagedDiff = "diff --git a/file.go b/file.go\nindex abcdef1..2345678 100644\n--- a/file.go\n+++ b/file.go\n@@ -1,1 +1,2 @@\n+func main() {\n+  fmt.Println(\"Hello\")\n}\n"

// setupTestConfig creates a default config for testing purposes.
func setupTestConfig() config.Config {
	cfg := config.NewDefaultConfig()
	cfg.AI.Providers[config.Gemini] = config.ProviderConfig{
		APIKey: "test-api-key",
		Model:  "gemini-pro",
	}
	return *cfg
}

func TestBuildPrompt_Basic(t *testing.T) {
	cfg := setupTestConfig()
	provider := GeminiProvider{cfg: cfg}

	prompt, err := provider.buildPrompt(stagedDiff, "")
	if err != nil {
		t.Fatalf("buildPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, stagedDiff) {
		t.Errorf("prompt missing staged diff")
	}
	if !strings.Contains(prompt, cfg.DefaultType) {
		t.Errorf("prompt missing default commit type")
	}
	for commitType := range cfg.Prompt.CommitTypes {
		if !strings.Contains(prompt, commitType) {
			t.Errorf("prompt missing commit type: %s", commitType)
		}
	}
}

func TestBuildPrompt_ForcedCommitType(t *testing.T) {
	cfg := setupTestConfig()
	cfg.ForcedCommitType = "feat"
	provider := GeminiProvider{cfg: cfg}

	prompt, err := provider.buildPrompt(stagedDiff, "")
	if err != nil {
		t.Fatalf("buildPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "- You MUST use the commit type: feat") {
		t.Errorf("prompt missing forced commit type instruction")
	}
	// Ensure the "Commit Types:" section is not present when forced
	if strings.Contains(prompt, "Commit Types:") {
		t.Errorf("prompt contains 'Commit Types:' section when forced commit type is set")
	}
}

func TestBuildPrompt_CustomCommitTypes(t *testing.T) {
	cfg := setupTestConfig()
	cfg.Prompt.CommitTypes = map[string]string{
		"custom": "A custom change",
		"new":    "A new entry",
	}
	provider := GeminiProvider{cfg: cfg}

	prompt, err := provider.buildPrompt(stagedDiff, "")
	if err != nil {
		t.Fatalf("buildPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "- custom: A custom change") {
		t.Errorf("prompt missing custom commit type 'custom'")
	}
	if !strings.Contains(prompt, "- new: A new entry") {
		t.Errorf("prompt missing custom commit type 'new'")
	}
}

func TestBuildPrompt_EmptyStagedDiff(t *testing.T) {
	cfg := setupTestConfig()
	provider := GeminiProvider{cfg: cfg}
	stagedDiff := ""

	prompt, err := provider.buildPrompt(stagedDiff, "")
	if err != nil {
		t.Fatalf("buildPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "```diff\n\n```") {
		t.Errorf("prompt does not correctly handle empty staged diff")
	}
}

func TestBuildPrompt_InvalidTemplate(t *testing.T) {
	cfg := setupTestConfig()
	cfg.Prompt.Template = "{{.StagedDiff" // Malformed template: Missing closing '}}'
	provider := GeminiProvider{cfg: cfg}

	_, err := provider.buildPrompt(stagedDiff, "")
	if err == nil {
		t.Errorf("expected an error for invalid template, got nil")
	}
	if !strings.Contains(err.Error(), "unclosed action") {
		t.Errorf("expected template parsing error for unclosed action, got: %v", err)
	}
}

func TestGenerate_WithEnvVar(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Warning: GEMINI_API_KEY not set, skipping TestGenerate_WithEnvVar. Please set the environment variable to run this test.")
	}

	cfg := setupTestConfig()
	cfg.AI.Providers[config.Gemini] = config.ProviderConfig{
		APIKey: apiKey,
		Model:  "gemini-2.5-flash", // Use a real model
	}
	cfg.SetupLocalProviderOverrides()

	provider, err := NewGeminiProvider(cfg)
	if err != nil {
		t.Fatalf("NewGeminiProvider failed: %v", err)
	}

	stagedDiff := "diff --git a/main.go b/main.go\nindex 0000000..abcdef0 100644\n--- a/main.go\n+++ b/main.go\n@@ -1,3 +1,7 @@\n package main\n \n import (\n+\t\"fmt\"\n \t\"log\"\n )\n \n+func greet() {\n+\tfmt.Println(\"Hello, world!\")\n}\n+\n func main() {\n \tlog.Println(\"Starting application\")\n+\tgreet()\n }"

	message, err := provider.Generate(context.Background(), stagedDiff, "")
	if err != nil {
		t.Fatalf("Generate test failed: %v", err)
	}

	if message == "" {
		t.Errorf("expected a non-empty commit message, got empty")
	}

	// Basic check for conventional commit format
	if !strings.Contains(message, ":") || !strings.Contains(message, "\n\n") {
		t.Logf("Generated message: %s", message)
		t.Errorf("generated message does not seem to follow conventional commit format")
	}
}

func TestBuildPrompt_ExistingCommitMessage(t *testing.T) {
	cfg := setupTestConfig()
	provider := GeminiProvider{cfg: cfg}
	existingMsg := "feat: existing feature\n\nThis is an existing message."

	prompt, err := provider.buildPrompt(stagedDiff, existingMsg)
	if err != nil {
		t.Fatalf("buildPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, existingMsg) {
		t.Errorf("prompt missing existing commit message")
	}
}
