package ai

import (
	"CommitGen/internal/config"
	"context"
)

// PromptData holds the necessary information to construct a commit message prompt for the LLM.
type PromptData struct {
	StagedDiff            string
	CommitTypes           map[string]string
	DefaultCommitType     string
	ForcedCommitType      string
	ExistingCommitMessage string
}

// LLMProvider defines the interface that large language model (LLM) providers must implement to generate commit messages.
type LLMProvider interface {
	buildPrompt(stagedDiff string, existingCommitMessage string) (string, error)
	Generate(ctx context.Context, stagedDiff string, existingCommitMessage string) (string, error)
}

// GetProvider returns an initialized LLMProvider implementation based on the configured default AI provider.
func GetProvider(cfg *config.Config) (LLMProvider, error) {
	switch cfg.AI.DefaultProvider {
	case config.Gemini:
		return NewGeminiProvider(cfg)
	case config.OpenAI:
		return nil, nil // TODO: add OpenAI support
	}
	return nil, nil
}
