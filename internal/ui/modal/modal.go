package modal

import tea "github.com/charmbracelet/bubbletea"

// Modal interface for all modals
type Modal interface {
	tea.Model
	ViewSized(width, height int) string
	IsOpen() bool
}
