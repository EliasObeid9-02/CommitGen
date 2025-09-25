package ai

import (
	"CommitGen/internal/config"
	"bytes"
	"context"
	"text/template"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	providerCfg config.ProviderConfig
	promptCfg   config.Prompt
	client      *genai.Client
}

type PromptData struct {
	StagedDiff  string
	CommitTypes map[string]string
}

func NewGeminiProvider(cfg config.Config) (*GeminiProvider, error) {
	providerCfg := cfg.AI.Providers[config.Gemini]
	promptCfg := cfg.Prompt

	client, err := genai.NewClient(context.TODO(), &genai.ClientConfig{
		APIKey:  providerCfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	provider := &GeminiProvider{
		providerCfg: providerCfg,
		promptCfg:   promptCfg,
		client:      client,
	}
	return provider, nil
}

func (p GeminiProvider) buildPrompt(stagedDiff string) (string, error) {
	data := PromptData{
		StagedDiff:  stagedDiff,
		CommitTypes: p.promptCfg.CommitTypes,
	}

	tmpl, err := template.New("prompt").Parse(p.promptCfg.Template)
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

	result, err := p.client.Models.GenerateContent(
		ctx,
		p.providerCfg.Model,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			Temperature:     &p.providerCfg.Temperature,
			MaxOutputTokens: p.providerCfg.MaxTokens,
		})

	if err != nil {
		return "", err
	}
	return result.Text(), nil
}
