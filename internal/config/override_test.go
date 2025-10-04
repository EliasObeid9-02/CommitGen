package config

import (
	"flag"
	"os"
	"testing"
)

/*
setupTestFlags prepares the flag package for isolated testing.
It saves and restores os.Args, and resets flag.CommandLine.
*/
func setupTestFlags(t *testing.T, args []string) (
	provider *string,
	apiKey *string,
	temperature *float64,
	model *string,
	maxTokens *int,
	commitType *string,
) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = append([]string{os.Args[0]}, args...)

	provider = flag.String("provider", "", "AI provider to use (e.g., gemini, openai)")
	apiKey = flag.String("api-key", "", "API key for the AI provider")
	temperature = flag.Float64("temperature", -1.0, "Temperature for the AI model")
	model = flag.String("model", "", "AI model to use")
	maxTokens = flag.Int("max-tokens", -1, "Maximum number of tokens for the AI model")
	commitType = flag.String("commit-type", "", "Type of commit (e.g., feat, fix, test)")
	flag.Parse()
	return
}

func TestOverrideFromFlags_ForcedCommitType(t *testing.T) {
	provider, apiKey, temperature, model, maxTokens, commitType := setupTestFlags(t, []string{"-commit-type", "feat"})

	cfg := NewDefaultConfig()
	cfg.OverrideFromFlags(commitType, provider, apiKey, model, temperature, maxTokens)

	if cfg.ForcedCommitType != "feat" {
		t.Errorf("expected ForcedCommitType 'feat', got %q", cfg.ForcedCommitType)
	}
}

func TestOverrideFromFlags_AIProviderSettings(t *testing.T) {
	provider, apiKey, temperature, model, maxTokens, commitType := setupTestFlags(t, []string{
		"-api-key", "test-key",
		"-model", "test-model",
		"-max-tokens", "500",
		"-temperature", "0.8",
	})

	cfg := NewDefaultConfig()
	cfg.OverrideFromFlags(commitType, provider, apiKey, model, temperature, maxTokens)

	providerCfg := cfg.AI.Providers[cfg.AI.DefaultProvider]
	if providerCfg.APIKey != "test-key" {
		t.Errorf("expected APIKey 'test-key', got %q", providerCfg.APIKey)
	}
	if providerCfg.Model != "test-model" {
		t.Errorf("expected Model 'test-model', got %q", providerCfg.Model)
	}
	if *providerCfg.MaxTokens != 500 {
		t.Errorf("expected MaxTokens 500, got %d", *providerCfg.MaxTokens)
	}
	if *providerCfg.Temperature != 0.8 {
		t.Errorf("expected Temperature 0.8, got %f", *providerCfg.Temperature)
	}
}

func TestOverrideFromFlags_SpecificProviderSettings(t *testing.T) {
	provider, apiKey, temperature, model, maxTokens, commitType := setupTestFlags(t, []string{
		"-provider", "gemini",
		"-api-key", "gemini-key",
		"-model", "gemini-model",
	})

	cfg := NewDefaultConfig()
	// Ensure Gemini provider exists in default config for this test
	if _, ok := cfg.AI.Providers[Gemini]; !ok {
		t.Fatalf("Gemini provider not found in default config, cannot test specific override.")
	}
	cfg.OverrideFromFlags(commitType, provider, apiKey, model, temperature, maxTokens)

	geminiCfg := cfg.AI.Providers[Gemini]
	if geminiCfg.APIKey != "gemini-key" {
		t.Errorf("expected Gemini APIKey 'gemini-key', got %q", geminiCfg.APIKey)
	}
	if geminiCfg.Model != "gemini-model" {
		t.Errorf("expected Gemini Model 'gemini-model', got %q", geminiCfg.Model)
	}
}

func TestOverrideFromFlags_NoFlags(t *testing.T) {
	provider, apiKey, temperature, model, maxTokens, commitType := setupTestFlags(t, []string{})

	initialCfg := NewDefaultConfig()
	cfg := NewDefaultConfig() // Create a separate config to modify
	cfg.OverrideFromFlags(commitType, provider, apiKey, model, temperature, maxTokens)

	/*
		Deep compare initialCfg and cfg to ensure no changes
		For simplicity, we'll check a few key fields. A more robust test might use reflection or a custom comparison function.
	*/
	if cfg.DefaultType != initialCfg.DefaultType {
		t.Errorf("DefaultType changed from %q to %q", initialCfg.DefaultType, cfg.DefaultType)
	}
	if cfg.AI.MaxTokens != initialCfg.AI.MaxTokens {
		t.Errorf("AI.MaxTokens changed from %d to %d", initialCfg.AI.MaxTokens, cfg.AI.MaxTokens)
	}
	// Check a provider setting
	initialGeminiCfg := initialCfg.AI.Providers[Gemini]
	currentGeminiCfg := cfg.AI.Providers[Gemini]
	if initialGeminiCfg.APIKey != currentGeminiCfg.APIKey {
		t.Errorf("Gemini APIKey changed from %q to %q", initialGeminiCfg.APIKey, currentGeminiCfg.APIKey)
	}
}

func TestOverrideFromFlags_PartialFlags(t *testing.T) {
	provider, apiKey, temperature, model, maxTokens, commitType := setupTestFlags(t, []string{"-api-key", "partial-key"})

	cfg := NewDefaultConfig()
	originalModel := cfg.AI.Providers[cfg.AI.DefaultProvider].Model // Store original model
	cfg.OverrideFromFlags(commitType, provider, apiKey, model, temperature, maxTokens)

	providerCfg := cfg.AI.Providers[cfg.AI.DefaultProvider]

	if providerCfg.APIKey != "partial-key" {
		t.Errorf("expected APIKey 'partial-key', got %q", providerCfg.APIKey)
	}
	// Ensure other fields remain unchanged
	if providerCfg.Model != originalModel {
		t.Errorf("expected Model to remain %q, got %q", originalModel, providerCfg.Model)
	}
}
