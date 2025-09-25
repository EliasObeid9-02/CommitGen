package ai

import (
	"CommitGen/internal/config"
	"context"
)

type LLMProvider interface {
	buildPrompt(stagedDiff string) (string, error)
	Generate(ctx context.Context, stagedDiff string) (string, error)
}

func GetProvider(cfg config.Config) (LLMProvider, error) {
	switch cfg.AI.DefaultProvider {
	case config.Gemini:
		return NewGeminiProvider(cfg)
	case config.OpenAI:
		return nil, nil // TODO: add OpenAI support
	}
	return nil, nil
}
