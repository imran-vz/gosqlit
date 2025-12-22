package styles

import "github.com/charmbracelet/lipgloss"

// Theme holds color scheme
type Theme struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Background lipgloss.Color
	Text       lipgloss.Color
	Border     lipgloss.Color
	Error      lipgloss.Color
	Success    lipgloss.Color
	Focused    lipgloss.Color
}

// DefaultTheme is the default color scheme
var DefaultTheme = Theme{
	Primary:    lipgloss.Color("62"),  // purple
	Secondary:  lipgloss.Color("42"),  // green
	Background: lipgloss.Color("235"), // dark gray
	Text:       lipgloss.Color("252"), // light gray
	Border:     lipgloss.Color("240"), // gray
	Error:      lipgloss.Color("196"), // red
	Success:    lipgloss.Color("46"),  // bright green
	Focused:    lipgloss.Color("63"),  // bright purple
}

// ActiveTheme is the currently active theme
var ActiveTheme = DefaultTheme

// Common styles
var (
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ActiveTheme.Border)

	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ActiveTheme.Focused)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ActiveTheme.Primary).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ActiveTheme.Error).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ActiveTheme.Success).
			Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ActiveTheme.Text).
			Background(ActiveTheme.Background)
)
