package ai

import (
	"CommitGen/internal/config"
	"testing"
)

func TestGeminiProvider_buildPrompt(t *testing.T) {
	// Create a mock configuration for testing
	cfg := config.Config{
		Prompt: config.Prompt{
			Template: "Staged Diff: {{.StagedDiff}} | Commit Types: {{range $key, $value := .CommitTypes}}{{$key}}-{{$value}},{{end}}",
			CommitTypes: map[string]string{
				"feat": "A new feature",
				"fix":  "A bug fix",
			},
		},
	}

	// Create a new GeminiProvider with the mock configuration
	provider := &GeminiProvider{
		promptCfg: cfg.Prompt,
	}

	// Define the test input
	stagedDiff := "this is a test diff"

	// Call the buildPrompt method
	prompt, err := provider.buildPrompt(stagedDiff)
	if err != nil {
		t.Fatalf("buildPrompt() error = %v, wantErr %v", err, false)
	}

	// Define the expected output
	expectedPrompt := "Staged Diff: this is a test diff | Commit Types: feat-A new feature,fix-A bug fix,"

	// Compare the actual and expected output
	if prompt != expectedPrompt {
		t.Errorf("buildPrompt() = %v, want %v", prompt, expectedPrompt)
	}
}