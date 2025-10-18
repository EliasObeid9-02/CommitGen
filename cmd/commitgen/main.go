package main

import (
	"flag"
	"path/filepath"
	"runtime"

	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const appName = "commitgen"

/*
getAppLoggerPath returns the path used for logging

It first checks for the XDG_STATE_HOME environment variable. If the variable
doesn't exist, it falls back based on the operating system.

The logger path falls back to on the of the following:

- Linux/macOS Fallback: ~/.local/state/commitgen/commitgen.log

- Windows Fallback:     %LOCALAPPDATA%\commitgen\commitgen.log
*/
func getAppLoggerPath() (string, error) {
	var stateDir string

	if xdgState := os.Getenv("XDG_STATE_HOME"); xdgState != "" {
		stateDir = xdgState
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not find user home directory: %w", err)
		}

		if runtime.GOOS == "windows" {
			stateDir = os.Getenv("LOCALAPPDATA")
			if stateDir == "" {
				stateDir = filepath.Join(homeDir, "AppData", "Local")
			}
		} else {
			stateDir = filepath.Join(homeDir, ".local", "state")
		}
	}

	logDir := filepath.Join(stateDir, appName)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", fmt.Errorf("could not create log directory at %s: %w", logDir, err)
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", appName))
	return logPath, nil
}

func main() {
	logFilePath, err := getAppLoggerPath()
	if err != nil {
		log.Fatalf("Error finding log file path: %v", err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer func() {
		err := logFile.Close()
		if err != nil {
			log.Fatalf("Error closing log file: %v", err)
		}
	}()

	logger := log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	// commitMsgFile := flag.String("commit-msg-file", "", "Path to the commit message file (used by git hook)")
	provider := flag.String("provider", "", "AI provider to use (e.g., gemini, openai)")
	apiKey := flag.String("api-key", "", "API key for the AI provider")
	model := flag.String("model", "", "AI model to use")
	commitType := flag.String("commit-type", "", "Type of commit (e.g., feat, fix, test)")
	temperature := flag.Float64("temperature", -1.0, "Temperature for the AI model")
	maxTokens := flag.Int("max-tokens", -1, "Maximum number of tokens for the AI model")
	flag.Parse()

	// After parsing flags, check for subcommands
	if len(flag.Args()) > 0 {
		switch flag.Args()[0] {
		case "install-hook":
			InstallHookFunc()
			return
		case "uninstall-hook":
			UninstallHookFunc()
			return
		case "generate-config":
			GenerateConfigFunc()
			return
		case "help":
			fmt.Println("Available commands: install-hook, uninstall-hook, generate-config")
			return
		}
	}

	app := initialApplication(logger, provider, apiKey, model, commitType, temperature, maxTokens)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Failed to start TUI application: %v", err)
		os.Exit(1)
	}

	// stagedDiff, err := git.GetStagedDiff()
	// if err != nil {
	// 	logger.Fatalf("Error getting stagedDiff: %v", err)
	// }
	//
	// var existingCommitMessage string
	// var existingCommitMessageComments string
	// if commitMsgFile != nil && *commitMsgFile != "" {
	// 	content, err := os.ReadFile(*commitMsgFile)
	// 	if err != nil {
	// 		logger.Printf("Warning: Could not read existing commit message file %s: %v", *commitMsgFile, err)
	// 	} else {
	// 		nonCommented, commented := git.ParseCommitMessage(string(content))
	// 		existingCommitMessage = nonCommented
	// 		existingCommitMessageComments = commented
	// 	}
	// }
	//
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()
	//
	// commitMessage, err := geminiProvider.Generate(ctx, stagedDiff, existingCommitMessage)
	// if err != nil {
	// 	logger.Fatalf("Error generating commit message: %v", err)
	// }
	//
	// finalCommitMessage := commitMessage
	// if existingCommitMessageComments != "" {
	// 	finalCommitMessage = fmt.Sprintf("%s\n%s", commitMessage, existingCommitMessageComments)
	// }
	//
	// if commitMsgFile != nil && *commitMsgFile != "" {
	// 	err := os.WriteFile(*commitMsgFile, []byte(finalCommitMessage), 0644)
	// 	if err != nil {
	// 		logger.Fatalf("Error writing commit message to file %s: %v", *commitMsgFile, err)
	// 	}
	// } else {
	// 	fmt.Println(finalCommitMessage)
	// }

}
