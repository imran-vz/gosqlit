package modal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/imran-vz/gosqlit/internal/config"
)

// ConnectionFormModal is form for adding/editing connections
type ConnectionFormModal struct {
	fields    []formField
	focusIdx  int
	isOpen    bool
	submitted bool
	isEdit    bool
	connID    string
}

type formField struct {
	label   string
	value   string
	masked  bool
	options []string // For driver selection
}

// NewConnectionForm creates connection form modal
func NewConnectionForm(existingConn *config.SavedConnection) *ConnectionFormModal {
	fields := []formField{
		{label: "Connection Name", value: "", masked: false},
		{label: "Driver", value: "postgres", masked: false, options: []string{"postgres"}},
		{label: "Host", value: "localhost", masked: false},
		{label: "Port", value: "5432", masked: false},
		{label: "Username", value: "", masked: false},
		{label: "Password", value: "", masked: true},
		{label: "Database", value: "", masked: false},
	}

	isEdit := false
	connID := uuid.New().String()

	if existingConn != nil {
		isEdit = true
		connID = existingConn.ID
		fields[0].value = existingConn.Name
		fields[1].value = existingConn.Driver
		fields[2].value = existingConn.Host
		fields[3].value = fmt.Sprintf("%d", existingConn.Port)
		fields[4].value = existingConn.Username
		fields[5].value = existingConn.Password
		fields[6].value = existingConn.Database
	}

	return &ConnectionFormModal{
		fields:   fields,
		focusIdx: 0,
		isOpen:   true,
		isEdit:   isEdit,
		connID:   connID,
	}
}

// Init initializes modal
func (cf *ConnectionFormModal) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (cf *ConnectionFormModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "esc":
			cf.isOpen = false
			return cf, nil
		case "tab", "down":
			cf.focusIdx = (cf.focusIdx + 1) % len(cf.fields)
		case "shift+tab", "up":
			cf.focusIdx--
			if cf.focusIdx < 0 {
				cf.focusIdx = len(cf.fields) - 1
			}
		case "left":
			if len(cf.fields[cf.focusIdx].value) > 0 {
				cf.fields[cf.focusIdx].value = cf.fields[cf.focusIdx].value[:len(cf.fields[cf.focusIdx].value)-1]
			}
		case "right":
			if len(cf.fields[cf.focusIdx].value) < len(cf.fields[cf.focusIdx].options) {
				cf.fields[cf.focusIdx].value = cf.fields[cf.focusIdx].options[len(cf.fields[cf.focusIdx].value)]
			}
		case "enter":
			// Save
			cf.submitted = true
			cf.isOpen = false
			return cf, nil
		case "backspace":
			if len(cf.fields[cf.focusIdx].value) > 0 {
				cf.fields[cf.focusIdx].value = cf.fields[cf.focusIdx].value[:len(cf.fields[cf.focusIdx].value)-1]
			}
		case "ctrl+u":
			// Clear field
			cf.fields[cf.focusIdx].value = ""
		case "ctrl+a":
			// Select all (just clear for now, easier to retype)
			cf.fields[cf.focusIdx].value = ""
		default:
			// Type characters (including paste support)
			input := keyMsg.String()

			// Strip bracketed paste mode markers
			input = stripPasteMarkers(input)

			// Filter out control characters but allow printable chars
			if len(input) > 0 && !isControlKey(input) {
				cf.fields[cf.focusIdx].value += input
			}
		}
	}

	return cf, nil
}

// View renders modal
func (cf *ConnectionFormModal) View() string {
	return cf.ViewSized(80, 24)
}

// ViewSized renders with specific dimensions
func (cf *ConnectionFormModal) ViewSized(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Padding(1, 0)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(20)

	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("237")).
		Padding(0, 1).
		Width(30)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60)

	title := "New Connection"
	if cf.isEdit {
		title = "Edit Connection"
	}

	content := titleStyle.Render(title) + "\n\n"

	for i, field := range cf.fields {
		label := labelStyle.Render(field.label + ":")

		value := field.value
		if field.masked && value != "" {
			value = maskString(value)
		}
		if value == "" {
			value = "____________"
		}

		input := inputStyle.Render(value)

		if i == cf.focusIdx {
			label = focusedStyle.Render("> " + field.label + ":")
			input = focusedStyle.Render(input)
		} else {
			label = "  " + label
		}

		content += label + " " + input + "\n"
	}

	content += "\n"
	content += lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Tab/↑↓: navigate  Ctrl+U: clear field  Enter: save  Esc: cancel")

	box := boxStyle.Render(content)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// IsOpen returns true if modal is open
func (cf *ConnectionFormModal) IsOpen() bool {
	return cf.isOpen
}

// GetConnection returns connection from form
func (cf *ConnectionFormModal) GetConnection() config.SavedConnection {
	port := 5432
	if cf.fields[3].value != "" {
		// Simple int parse
		p := 0
		for _, c := range cf.fields[3].value {
			if c >= '0' && c <= '9' {
				p = p*10 + int(c-'0')
			}
		}
		if p > 0 {
			port = p
		}
	}

	return config.SavedConnection{
		ID:       cf.connID,
		Name:     cf.fields[0].value,
		Driver:   cf.fields[1].value,
		Host:     cf.fields[2].value,
		Port:     port,
		Username: cf.fields[4].value,
		Password: cf.fields[5].value,
		Database: cf.fields[6].value,
	}
}

// IsSubmitted returns true if submitted
func (cf *ConnectionFormModal) IsSubmitted() bool {
	return cf.submitted
}

func maskString(s string) string {
	result := ""
	for range s {
		result += "*"
	}
	return result
}
