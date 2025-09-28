package ai

import (
	"CommitGen/internal/config"
	"bytes"
	"context"
	"fmt"
	"text/template"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	cfg    config.Config
	client *genai.Client
}

type PromptData struct {
	StagedDiff        string
	CommitTypes       map[string]string
	DefaultCommitType string
	ForcedCommitType  string
}

func NewGeminiProvider(cfg config.Config) (*GeminiProvider, error) {
	providerCfg := cfg.AI.Providers[config.Gemini]

	client, err := genai.NewClient(context.TODO(), &genai.ClientConfig{
		APIKey:  providerCfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	provider := &GeminiProvider{
		cfg:    cfg,
		client: client,
	}
	return provider, nil
}

func (p GeminiProvider) buildPrompt(stagedDiff string) (string, error) {
	data := PromptData{
		StagedDiff:        stagedDiff,
		CommitTypes:       p.cfg.Prompt.CommitTypes,
		DefaultCommitType: p.cfg.DefaultType,
		ForcedCommitType:  p.cfg.ForcedCommitType,
	}

	tmpl, err := template.New("prompt").Parse(p.cfg.Prompt.Template)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (p GeminiProvider) Generate(ctx context.Context, stagedDiff string) (string, error) {
	prompt, err := p.buildPrompt(stagedDiff)
	if err != nil {
		return "", err
	}

	providerCfg := p.cfg.AI.Providers[config.Gemini]
	result, err := p.client.Models.GenerateContent(
		ctx,
		providerCfg.Model,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			Temperature:     providerCfg.Temperature,
			MaxOutputTokens: *providerCfg.MaxTokens,
		})

	if err != nil {
		return "", err
	}

	if result == nil || len(result.Candidates) == 0 {
		return "", fmt.Errorf("received an empty response from the AI provider")
	}

	// Check the finish reason. If it's not 'STOP', the model was likely blocked.
	candidate := result.Candidates[0]
	if candidate.FinishReason != genai.FinishReasonStop {
		return "", fmt.Errorf("AI generation stopped for reason: %s", candidate.FinishReason)
	}

	var responseBuilder bytes.Buffer
	if candidate.Content != nil {
		for _, part := range candidate.Content.Parts {
			responseBuilder.WriteString(part.Text)
		}
	}

	if responseBuilder.Len() == 0 {
		return "", fmt.Errorf("AI returned a candidate with zero parts")
	}
	return responseBuilder.String(), nil
}
