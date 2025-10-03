package main

import (
	"CommitGen/internal/config"
	"CommitGen/internal/git"
	"fmt"
	"log"
)

// InstallHookFunc installs the git hook by calling git.Install function.
func InstallHookFunc() {
	err := git.Install()
	if err != nil {
		log.Fatalf("Error installing hook: %v", err)
	}
	fmt.Println("Git hook installed successfully.")
}

// UninstallHookFunc uninstalls the git hook by calling git.uninstall function.
func UninstallHookFunc() {
	err := git.Uninstall()
	if err != nil {
		log.Fatalf("Error uninstalling hook: %v", err)
	}
	fmt.Println("Git hook uninstalled successfully.")
}

// GenerateConfigFunc writes the default config file by calling config.GenerateConfig function.
func GenerateConfigFunc() {
	err := config.GenerateConfig()
	if err != nil {
		log.Fatalf("Error generating config: %v", err)
	}
	fmt.Println("Config file generated successfully.")
}
