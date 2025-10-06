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

	// ForcedCommitType is used to override the commit type from the command line.
	ForcedCommitType string `toml:"-"`
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
	APIKey      string   `toml:"api_key" comment:"Your secret API key for this provider."`
	Model       string   `toml:"model" comment:"The specific model to use (e.g., 'gemini-2.5-flash')."`
	MaxTokens   *int32   `toml:"max_tokens" comment:"Optional: Overrides the global max_tokens setting for this provider."`
	Temperature *float32 `toml:"temperature" comment:"Optional: Overrides the global temperature setting for this provider."`
}

// Prompt holds the prompt-related settings.
type Prompt struct {
	Template    string            `toml:"template,multiline" comment:"The prompt template. Use {{.StagedDiff}} for staged changes and {{.CommitTypes}} for the types list."`
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
		MaxTokens:       4096,
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
		Template: `You are an expert at writing conventional commit messages.

**INSTRUCTIONS:**
- Your primary task is to generate a Git commit message based on the provided staged diff.
- If an existing commit message is provided, amend it based on the new diff and instructions.
{{if .ForcedCommitType}}
- You MUST use the commit type: {{.ForcedCommitType}}
{{else}}
- Choose the best commit type from the provided list.
- If you are unsure which type to use, default to: {{.DefaultCommitType}}
{{end}}

**GUIDELINES:**
- Focus on explaining the 'why' of the changes, not just the 'what'.
- Aim for clarity, conciseness, and descriptiveness in the summary and body.
- Consider the broader context of the changes (e.g., feature, bug fix, refactor).

**RULES:**
- The commit message must be written in English.
- Do not include any conversational text, explanations, or meta-commentary outside of the commit message itself.
- Do not include sensitive information or personal opinions.
- Ensure the message accurately reflects the changes in the staged diff.

**FORMAT:**
{commit_type}{commit_scope (optional)}: {commit_summary}

{commit_body}

- The subject line (first line) must be 50-72 characters or less.
- The commit summary should start with a capital letter.
- Separate the subject line from the body with a blank line.
- Wrap body lines at 72 characters.
- The body should be a collection of bullet points explaining the details of the commit.
- Bullet points should uses dashes and not asterisks.
- The scope is optional and should be surrounded by parentheses.

{{if .ExistingCommitMessage}}
**EXISTING COMMIT MESSAGE:**{{.ExistingCommitMessage}}
{{end}}

**STAGED DIFF:**
{{.StagedDiff}}


{{if not .ForcedCommitType}}
**COMMIT TYPES:**
{{range $type, $description := .CommitTypes}}
- {{$type}}: {{$description}}
{{end}}
{{end}}

The final output should be only the raw commit message, without any markdown formatting.`,
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
