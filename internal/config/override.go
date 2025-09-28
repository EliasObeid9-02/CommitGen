package config

import (
	"flag"
)

// OverrideFromFlags modifies the configuration based on command-line flags.
func (c *Config) OverrideFromFlags() {
	provider := flag.String("provider", "", "AI provider to use (e.g., gemini, openai)")
	apiKey := flag.String("api-key", "", "API key for the AI provider")
	temperature := flag.Float64("temperature", -1.0, "Temperature for the AI model")
	model := flag.String("model", "", "AI model to use")
	maxTokens := flag.Int("max-tokens", -1, "Maximum number of tokens for the AI model")
	commitType := flag.String("commit-type", "", "Type of commit (e.g., feat, fix, test)")
	flag.Parse()

	if *commitType != "" {
		c.ForcedCommitType = *commitType
	}

	var targetProvider ProviderType
	if *provider != "" {
		targetProvider = ProviderType(*provider)
	} else {
		targetProvider = c.AI.DefaultProvider
	}

	providerConfig := c.AI.Providers[targetProvider]
	if *apiKey != "" {
		providerConfig.APIKey = *apiKey
	}
	if *model != "" {
		providerConfig.Model = *model
	}
	if *maxTokens != -1 {
		// Allocate new memory for the value and assign its address.
		// This prevents a dangling pointer to a local variable.
		newMaxTokens := int32(*maxTokens)
		providerConfig.MaxTokens = &newMaxTokens
	}
	if *temperature != -1.0 {
		// Allocate new memory for the value and assign its address.
		newTemp := float32(*temperature)
		providerConfig.Temperature = &newTemp
	}
	c.AI.Providers[targetProvider] = providerConfig
}
