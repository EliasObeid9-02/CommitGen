package main

import (
	"CommitGen/internal/ai"
	"CommitGen/internal/config"
	"CommitGen/internal/git"
	"context"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type application struct {
	logger   *log.Logger
	provider ai.LLMProvider
}

func initialApplication(logger *log.Logger, providerName, apiKey, model, commitType *string, temperature *float64, maxTokens *int) application {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Error loading configuration: %v", err)
	}
	cfg.OverrideFromFlags(commitType, providerName, apiKey, model, temperature, maxTokens)

	provider, err := ai.GetProvider(cfg)
	if err != nil {
		logger.Fatalf("Error initializing AI provider: %v", err)
	}

	return application{
		logger:   logger,
		provider: provider,
	}
}

func (a application) Init() tea.Cmd {
	return nil
}

func (a application) View() string {
	return ""
}

func (a application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

	case commitMessageMsg:

	case errorMsg:
		a.logger.Printf("Encounterd error: %v\n", msg.err)
	}
	return a, nil
}

type commitMessageMsg struct{ msg string }
type errorMsg struct{ err error }

func (a application) generateCommitMessageCmd() tea.Msg {
	stagedDiff, err := git.GetStagedDiff()
	if err != nil {
		return errorMsg{err}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	commitMsg, err := a.provider.Generate(ctx, stagedDiff, "")
	if err != nil {
		return errorMsg{err}
	}
	return commitMessageMsg{commitMsg}
}
