package main

import (
	"CommitGen/internal/ai"
	"CommitGen/internal/config"
	"CommitGen/internal/git"
	"context"
	"path/filepath"
	"runtime"
	"time"

	"fmt"
	"log"
	"os"
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

	if len(os.Args) > 1 {
		switch os.Args[1] {
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

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Error loading configuration: %v", err)
	}
	cfg.OverrideFromFlags()

	// Initialize the AI provider
	geminiProvider, err := ai.NewGeminiProvider(*cfg)
	if err != nil {
		logger.Fatalf("Error initializing AI provider: %v", err)
	}

	stagedDiff, err := git.GetStagedDiff()
	if err != nil {
		logger.Fatalf("Error getting stagedDiff: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	commitMessage, err := geminiProvider.Generate(ctx, stagedDiff)
	if err != nil {
		logger.Fatalf("Error generating commit message: %v", err)
	}
	fmt.Println(commitMessage)
}
