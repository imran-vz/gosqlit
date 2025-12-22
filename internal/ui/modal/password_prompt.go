package modal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PasswordPromptModal prompts for master password
type PasswordPromptModal struct {
	password      string
	confirmPass   string
	focusConfirm  bool   // true if focus on confirm field
	error         string
	submitted     bool
	isOpen        bool
	isNewSetup    bool // true if setting up password first time
}

// NewPasswordPrompt creates password prompt modal
func NewPasswordPrompt(isNewSetup bool) *PasswordPromptModal {
	return &PasswordPromptModal{
		isOpen:     true,
		isNewSetup: isNewSetup,
	}
}

// Init initializes the modal
func (p *PasswordPromptModal) Init() tea.Cmd {
	return nil
}

// Update handles messages (tea.Model interface)
func (p *PasswordPromptModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return p, tea.Quit
		case "tab":
			if p.isNewSetup {
				p.focusConfirm = !p.focusConfirm
			}
		case "enter":
			if p.isNewSetup {
				if p.password == "" {
					p.error = "Password cannot be empty"
					return p, nil
				}
				if !p.focusConfirm {
					// Move to confirm field
					p.focusConfirm = true
					return p, nil
				}
				if p.confirmPass == "" {
					p.error = "Please confirm password"
					return p, nil
				}
				if p.password != p.confirmPass {
					p.error = "Passwords do not match"
					p.password = ""
					p.confirmPass = ""
					p.focusConfirm = false
					return p, nil
				}
			} else {
				if p.password == "" {
					p.error = "Password cannot be empty"
					return p, nil
				}
			}
			p.submitted = true
			p.isOpen = false
			return p, tea.Quit  // Exit the Bubble Tea program!
		case "backspace":
			if p.isNewSetup && p.focusConfirm {
				if len(p.confirmPass) > 0 {
					p.confirmPass = p.confirmPass[:len(p.confirmPass)-1]
					p.error = ""
				}
			} else {
				if len(p.password) > 0 {
					p.password = p.password[:len(p.password)-1]
					p.error = ""
				}
			}
		case "esc":
			return p, tea.Quit
		default:
			// Add characters (including paste support)
			input := msg.String()
			// Filter out control characters but allow printable chars
			if len(input) > 0 && !isControlKey(input) {
				if p.isNewSetup && p.focusConfirm {
					p.confirmPass += input
				} else {
					p.password += input
				}
				p.error = ""
			}
		}
	}
	return p, nil
}

// View renders the modal (tea.Model interface)
func (p *PasswordPromptModal) View() string {
	return p.ViewSized(80, 24) // Default size
}

// ViewSized renders with specific dimensions
func (p *PasswordPromptModal) ViewSized(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Padding(1, 0)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Padding(0, 1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(50)

	var title string
	if p.isNewSetup {
		title = "Setup Master Password"
	} else {
		title = "Enter Master Password"
	}

	content := titleStyle.Render(title) + "\n\n"

	if p.isNewSetup {
		content += "This password will encrypt your saved connections.\n"
		content += "Make sure you remember it!\n\n"
	}

	// Show cursor indicator and instructions
	passPrompt := "Password: "
	if !p.isNewSetup || !p.focusConfirm {
		passPrompt = "> " + passPrompt
	} else {
		passPrompt = "  " + passPrompt
	}
	content += passPrompt + maskPassword(p.password) + "\n"

	if p.isNewSetup {
		confPrompt := "Confirm:  "
		if p.focusConfirm {
			confPrompt = "> " + confPrompt
		} else {
			confPrompt = "  " + confPrompt
		}
		content += confPrompt + maskPassword(p.confirmPass) + "\n"

		content += "\n"
		// Add very clear instructions
		if !p.focusConfirm && p.password == "" {
			content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Step 1: Enter your master password")
		} else if !p.focusConfirm && p.password != "" {
			content += lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true).
				Render("✓ Password entered! Press Enter to continue...")
		} else if p.focusConfirm && p.confirmPass == "" {
			content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Step 2: Type the password again to confirm")
		} else if p.focusConfirm && p.confirmPass != "" {
			content += lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true).
				Render("✓ Confirmed! Press Enter to save...")
		}
	}

	if p.error != "" {
		content += "\n" + errorStyle.Render(p.error)
	}

	content += "\n\n"
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to continue, Esc to quit")

	box := boxStyle.Render(content)

	// Center the box
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// IsOpen returns true if modal is open
func (p *PasswordPromptModal) IsOpen() bool {
	return p.isOpen
}

// GetPassword returns entered password
func (p *PasswordPromptModal) GetPassword() string {
	return p.password
}

// IsSubmitted returns true if user submitted
func (p *PasswordPromptModal) IsSubmitted() bool {
	return p.submitted
}

// SetError sets error message
func (p *PasswordPromptModal) SetError(err string) {
	p.error = err
	p.password = ""
	p.confirmPass = ""
	p.isOpen = true
	p.submitted = false
}

// maskPassword masks password with asterisks
func maskPassword(s string) string {
	if s == "" {
		return "____________"
	}
	masked := ""
	for range s {
		masked += "*"
	}
	return masked
}
