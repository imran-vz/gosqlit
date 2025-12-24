package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/imran-vz/gosqlit/internal/app"
	"github.com/imran-vz/gosqlit/internal/config"
	"github.com/imran-vz/gosqlit/internal/debug"
	"github.com/imran-vz/gosqlit/internal/ui/modal"

	// Import drivers to register them
	_ "github.com/imran-vz/gosqlit/internal/db/drivers/postgres"
)

var (
	debugMode = flag.Bool("debug", false, "Enable debug mode")
	logFile   = flag.String("log", "", "Debug log file (default: stderr)")
)

func main() {
	flag.Parse()

	// Initialize debug mode
	if err := debug.Init(*debugMode, *logFile); err != nil {
		fmt.Printf("Failed to initialize debug mode: %v\n", err)
		os.Exit(1)
	}
	defer debug.Close()

	debug.Logf("Starting gosqlit with debug mode: %v", *debugMode)

	// Check if config exists
	tmpMgr, err := config.NewManager("")
	if err != nil {
		debug.LogError(err, "main/config_init")
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}

	isNewSetup := !tmpMgr.Exists()
	debug.Logf("Is new setup: %v", isNewSetup)

	// Show password prompt
	passwordPrompt := modal.NewPasswordPrompt(isNewSetup)
	p := tea.NewProgram(passwordPrompt)

	result, err := p.Run()
	if err != nil {
		debug.LogError(err, "main/password_prompt")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get password from modal
	promptResult := result.(*modal.PasswordPromptModal)
	if !promptResult.IsSubmitted() {
		debug.Log("User cancelled password prompt")
		fmt.Println("Cancelled")
		os.Exit(0)
	}

	password := promptResult.GetPassword()
	debug.Log("Master password provided")

	// Initialize config manager with password
	configMgr, err := config.NewManager(password)
	if err != nil {
		debug.LogError(err, "main/config_manager")
		fmt.Printf("Error initializing config manager: %v\n", err)
		os.Exit(1)
	}

	// Load config
	cfg, err := configMgr.Load()
	if err != nil {
		debug.LogError(err, "main/config_load")
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("This might mean your password is incorrect.")
		os.Exit(1)
	}

	// If new setup, save empty config to create file
	if isNewSetup {
		if err := configMgr.Save(cfg); err != nil {
			debug.LogError(err, "main/config_save")
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}
		debug.Log("New setup completed successfully")
		fmt.Println("Master password set successfully!")
	}

	// Create and run main app
	application := app.New(configMgr)
	mainProgram := tea.NewProgram(application, tea.WithAltScreen())

	debug.Log("Starting main application")
	if _, err := mainProgram.Run(); err != nil {
		debug.LogError(err, "main/app_run")
		fmt.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}

	debug.Log("Application ended normally")
}
