package connected

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar displays connection and query info
type StatusBar struct {
	connInfo     string
	queryTime    time.Duration
	rowCount     int
	errorMsg     string
	queryRunning bool
	width        int
}

// NewStatusBar creates status bar
func NewStatusBar(connInfo string) *StatusBar {
	return &StatusBar{
		connInfo: connInfo,
	}
}

// View renders the status bar
func (sb *StatusBar) View() string {
	leftStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("235")).
		Padding(0, 2)

	rightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("235")).
		Padding(0, 2)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Background(lipgloss.Color("235")).
		Bold(true).
		Padding(0, 2)

	left := sb.connInfo
	right := ""

	if sb.errorMsg != "" {
		right = errorStyle.Render("Error: " + sb.errorMsg)
	} else if sb.queryRunning {
		right = rightStyle.Render("⏳ Running... (Ctrl+K to cancel)")
	} else if sb.queryTime > 0 {
		right = rightStyle.Render(fmt.Sprintf("✓ %d rows in %v", sb.rowCount, sb.queryTime))
	}

	leftRendered := leftStyle.Render(left)
	rightRendered := right

	// Calculate spacing
	leftWidth := lipgloss.Width(leftRendered)
	rightWidth := lipgloss.Width(rightRendered)
	spacing := sb.width - leftWidth - rightWidth

	if spacing < 0 {
		spacing = 0
	}

	spacer := strings.Repeat(" ", spacing)
	spacerStyle := lipgloss.NewStyle().Background(lipgloss.Color("235"))

	return leftRendered + spacerStyle.Render(spacer) + rightRendered
}

// SetQueryResult sets successful query result
func (sb *StatusBar) SetQueryResult(rowCount int, elapsed time.Duration) {
	sb.rowCount = rowCount
	sb.queryTime = elapsed
	sb.errorMsg = ""
	sb.queryRunning = false
}

// SetError sets error message
func (sb *StatusBar) SetError(err string) {
	sb.errorMsg = err
	sb.queryRunning = false
	sb.queryTime = 0
	sb.rowCount = 0
}

// SetQueryRunning sets running state
func (sb *StatusBar) SetQueryRunning(running bool) {
	sb.queryRunning = running
	if running {
		sb.errorMsg = ""
	}
}

// SetWidth sets status bar width
func (sb *StatusBar) SetWidth(width int) {
	sb.width = width
}
