package main

import (
	"CommitGen/internal/config"
	"CommitGen/internal/git"
	"fmt"
	"log"
)

func InstallHookFunc() {
	err := git.Install()
	if err != nil {
		log.Fatalf("Error installing hook: %v", err)
	}
	fmt.Println("Git hook installed successfully.")
}

func UninstallHookFunc() {
	err := git.Uninstall()
	if err != nil {
		log.Fatalf("Error uninstalling hook: %v", err)
	}
	fmt.Println("Git hook uninstalled successfully.")
}

func GenerateConfigFunc() {
	err := config.GenerateConfig()
	if err != nil {
		log.Fatalf("Error generating config: %v", err)
	}
	fmt.Println("Config file generated successfully.")
}
