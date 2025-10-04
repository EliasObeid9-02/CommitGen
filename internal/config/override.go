package config

// OverrideFromFlags modifies the configuration based on command-line flags.
func (c *Config) OverrideFromFlags(
	commitType,
	provider,
	apiKey,
	model *string,
	temperature *float64,
	maxTokens *int,

) {
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
