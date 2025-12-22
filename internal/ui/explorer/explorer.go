package explorer

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/imran-vz/gosqlit/internal/config"
	"github.com/imran-vz/gosqlit/internal/ui/styles"
)

// ExplorerView shows saved connections
type ExplorerView struct {
	connections []config.SavedConnection
	cursor      int
	scroll      int
	width       int
	height      int
}

// NewExplorer creates explorer view
func NewExplorer(connections []config.SavedConnection) *ExplorerView {
	return &ExplorerView{
		connections: connections,
		cursor:      0,
		scroll:      0,
	}
}

// Update handles messages
func (e *ExplorerView) Update(msg tea.Msg) (*ExplorerView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.width = msg.Width
		e.height = msg.Height
		return e, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if e.cursor > 0 {
				e.cursor--
				// Scroll up if needed
				if e.cursor < e.scroll {
					e.scroll = e.cursor
				}
			}
		case "down", "j":
			if e.cursor < len(e.connections)-1 {
				e.cursor++
				// Scroll down if needed
				maxVisible := e.height - 10 // Reserve space for header/footer
				if e.cursor >= e.scroll+maxVisible {
					e.scroll++
				}
			}
		case "home":
			e.cursor = 0
			e.scroll = 0
		case "end":
			e.cursor = len(e.connections) - 1
			maxVisible := e.height - 10
			if len(e.connections) > maxVisible {
				e.scroll = len(e.connections) - maxVisible
			}
		}
	}

	return e, nil
}

// View renders the view
func (e *ExplorerView) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.ActiveTheme.Primary).
		Bold(true).
		Padding(1, 2)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(1, 2)

	itemStyle := lipgloss.NewStyle().
		Padding(0, 2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.ActiveTheme.Focused).
		Background(lipgloss.Color("237")).
		Bold(true).
		Padding(0, 2)

	var content string

	// Header
	content += titleStyle.Render("GosQLit - Database Connections") + "\n\n"

	if len(e.connections) == 0 {
		content += itemStyle.Render("No saved connections.") + "\n\n"
		content += itemStyle.Render("Press 'n' to create a new connection.") + "\n"
	} else {
		// List connections
		maxVisible := e.height - 10
		start := e.scroll
		end := e.scroll + maxVisible
		if end > len(e.connections) {
			end = len(e.connections)
		}

		for i := start; i < end; i++ {
			conn := e.connections[i]

			cursor := "  "
			if i == e.cursor {
				cursor = "> "
			}

			line := fmt.Sprintf("%s%-25s  %s @ %s:%d",
				cursor,
				conn.Name,
				conn.Driver,
				conn.Host,
				conn.Port,
			)

			if i == e.cursor {
				content += selectedStyle.Render(line) + "\n"
			} else {
				content += itemStyle.Render(line) + "\n"
			}
		}

		// Show scroll indicator
		if len(e.connections) > maxVisible {
			content += "\n" + helpStyle.Render(fmt.Sprintf(
				"Showing %d-%d of %d",
				start+1,
				end,
				len(e.connections),
			))
		}
	}

	// Help
	content += "\n\n"
	content += helpStyle.Render("↑/↓: navigate  Enter: connect  n: new  d: delete  e: edit  q: quit")

	// Center content
	box := lipgloss.NewStyle().
		Width(e.width).
		Height(e.height).
		Align(lipgloss.Left, lipgloss.Top).
		Render(content)

	return box
}

// GetSelectedConnection returns currently selected connection
func (e *ExplorerView) GetSelectedConnection() *config.SavedConnection {
	if e.cursor >= 0 && e.cursor < len(e.connections) {
		return &e.connections[e.cursor]
	}
	return nil
}

// SetConnections updates connection list
func (e *ExplorerView) SetConnections(connections []config.SavedConnection) {
	e.connections = connections
	if e.cursor >= len(e.connections) {
		e.cursor = len(e.connections) - 1
	}
	if e.cursor < 0 {
		e.cursor = 0
	}
}
