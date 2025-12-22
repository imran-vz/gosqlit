package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/imran-vz/gosqlit/internal/app"
	"github.com/imran-vz/gosqlit/internal/config"
	"github.com/imran-vz/gosqlit/internal/ui/modal"

	// Import drivers to register them
	_ "github.com/imran-vz/gosqlit/internal/db/drivers/postgres"
)

func main() {
	// Check if config exists
	tmpMgr, err := config.NewManager("")
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}

	isNewSetup := !tmpMgr.Exists()

	// Show password prompt
	passwordPrompt := modal.NewPasswordPrompt(isNewSetup)
	p := tea.NewProgram(passwordPrompt)

	result, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get password from modal
	promptResult := result.(*modal.PasswordPromptModal)
	if !promptResult.IsSubmitted() {
		fmt.Println("Cancelled")
		os.Exit(0)
	}

	password := promptResult.GetPassword()

	// Initialize config manager with password
	configMgr, err := config.NewManager(password)
	if err != nil {
		fmt.Printf("Error initializing config manager: %v\n", err)
		os.Exit(1)
	}

	// Load config
	cfg, err := configMgr.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("This might mean your password is incorrect.")
		os.Exit(1)
	}

	// If new setup, save empty config to create file
	if isNewSetup {
		if err := configMgr.Save(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Master password set successfully!")
	}

	// Create and run main app
	application := app.New(configMgr)
	mainProgram := tea.NewProgram(application, tea.WithAltScreen())

	if _, err := mainProgram.Run(); err != nil {
		fmt.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}
}
