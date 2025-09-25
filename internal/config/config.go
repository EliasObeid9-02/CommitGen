package config

import (
	"os"
)

// Config holds the application-wide settings, loaded from a TOML file.
type Config struct {
	// General settings for the application.
	DefaultType     string `toml:"default_type" comment:"The default commit type if no flag is provided (e.g., 'feat')."`
	Editor          string `toml:"editor" comment:"The preferred text editor for editing the commit message. Overrides $EDITOR and $VISUAL."`
	CommitUserName  string `toml:"commit_username" comment:"Optional: Overrides the system Git user name."`
	CommitUserEmail string `toml:"commit_email" comment:"Optional: Overrides the system Git user email."`

	// AI is a table for AI-related configuration.
	AI AI `toml:"ai"`

	// Prompt is a table for prompt-related configuration.
	Prompt Prompt `toml:"prompt"`
}

// AI holds global and provider-specific settings for the AI service.
type AI struct {
	DefaultProvider ProviderType `toml:"default_provider" comment:"The default AI service to use (e.g., 'gemini'). Must match a provider key below."`
	MaxTokens       int32        `toml:"max_tokens" comment:"Global default for the maximum number of tokens for the generated response."`
	Temperature     float32      `toml:"temperature" comment:"Global default between 0.0 and 1.0 that controls the randomness of the AI's output. Lower is more predictable."`
	Providers       ProviderMap  `toml:"providers" comment:"Configurations for each AI provider."`
}

// ProviderMap maps ProviderType to their corresponding configs
type ProviderMap map[ProviderType]ProviderConfig

// ProviderType is a custom type to ensure only supported provider names are used.
type ProviderType string

const (
	Gemini ProviderType = "gemini"
	OpenAI ProviderType = "openai"
)

// ProviderConfig holds the specific settings for a single AI provider.
type ProviderConfig struct {
	APIKey      string  `toml:"api_key" comment:"Your secret API key for this provider."`
	Model       string  `toml:"model" comment:"The specific model to use (e.g., 'gemini-2.5-flash')."`
	MaxTokens   int32   `toml:"max_tokens" comment:"Optional: Overrides the global max_tokens setting for this provider."`
	Temperature float32 `toml:"temperature" comment:"Optional: Overrides the global temperature setting for this provider."`
}

// Prompt holds the prompt-related settings.
type Prompt struct {
	Template    string            `toml:"template" comment:"The prompt template. Use {{.StagedDiff}} for staged changes and {{.CommitTypes}} for the types list."`
	CommitTypes map[string]string `toml:"commit_types" comment:"A map of commit types and their descriptions for the AI to choose from."`
}

// NewDefaultConfig returns a Config struct with all default values.
func NewDefaultConfig() *Config {
	return &Config{
		DefaultType:     "refactor",
		Editor:          findEditor(),
		CommitUserEmail: "",
		CommitUserName:  "",
		AI:              NewDefaultAIConfig(),
		Prompt:          NewDefaultPromptConfig(),
	}
}

// NewDefaultAIConfig creates the default AI configuration.
func NewDefaultAIConfig() AI {
	return AI{
		DefaultProvider: Gemini,
		MaxTokens:       256,
		Temperature:     0.3,
		Providers: map[ProviderType]ProviderConfig{
			Gemini: {
				APIKey: "",
				Model:  "gemini-2.5-flash",
			},
		},
	}
}

// NewDefaultPromptConfig creates the default prompt configuration.
func NewDefaultPromptConfig() Prompt {
	return Prompt{
		Template: `
You are an Senior Software Engineer with years of experience in writing concise and conventional commit messages.
Based on the following staged diff, please generate a commit message.

**Staged Diff:**
` + "```" + `diff
{{.StagedDiff}}
` + "```" + `

**Commit Types:**
{{range $type, $description := .CommitTypes}}
- {{$type}}: {{$description}}
{{end}}

Please follow the conventional commit format. The final output should be only the commit message.
`,
		CommitTypes: map[string]string{
			"feat":     "A new feature",
			"fix":      "A bug fix",
			"docs":     "Documentation only changes",
			"style":    "Changes that do not affect the meaning of the code",
			"refactor": "A code change that neither adds a feature nor fixes a bug",
			"perf":     "A code change that improves performance",
			"test":     "Adding missing tests or correcting existing tests",
			"chore":    "Changes to the build process or auxiliary tools",
			"build":    "Changes that affect the build system or external dependencies",
			"ci":       "Changes to your Continuous Integration configuration",
		},
	}
}

// findEditor determines the default text editor by checking environment variables
// in the order of $VISUAL, then $EDITOR. It defaults to "nano" if neither is set.
func findEditor() string {
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "nano"
}
