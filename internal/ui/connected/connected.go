package connected

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PaneType represents which pane is focused
type PaneType int

const (
	PaneSchemaBrowser PaneType = iota
	PaneEditor
	PaneResults
)

// ConnectedView is main workspace view
type ConnectedView struct {
	ConnID       string
	FocusedPane  PaneType
	LeftWidth    int // percentage (0-100)
	QueryRunning bool
	CancelFunc   context.CancelFunc

	// Components
	Browser   *SchemaBrowser
	Editor    *QueryEditor
	Results   *ResultsTable
	StatusBar *StatusBar

	// Dimensions
	width  int
	height int
}

// NewConnectedView creates new connected view
func NewConnectedView(connID string, connInfo string) *ConnectedView {
	return &ConnectedView{
		ConnID:      connID,
		FocusedPane: PaneEditor,
		LeftWidth:   25, // 25% for left panel

		Browser:   NewSchemaBrowser(),
		Editor:    NewQueryEditor(),
		Results:   NewResultsTable(),
		StatusBar: NewStatusBar(connInfo),
	}
}

// Update handles messages
func (cv *ConnectedView) Update(msg tea.Msg) (*ConnectedView, tea.Cmd) {
	var cmd tea.Cmd

	// Handle window size
	if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
		cv.width = wsMsg.Width
		cv.height = wsMsg.Height
		cv.updateDimensions()
		return cv, nil
	}

	// Handle keyboard
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab":
			// Cycle focus
			switch cv.FocusedPane {
			case PaneSchemaBrowser:
				cv.FocusedPane = PaneEditor
			case PaneEditor:
				cv.FocusedPane = PaneResults
			case PaneResults:
				cv.FocusedPane = PaneSchemaBrowser
			}
			return cv, nil

		case "ctrl+enter":
			// Execute query
			// TODO: will be handled in app.go
			return cv, nil
		}
	}

	// Delegate to focused pane
	switch cv.FocusedPane {
	case PaneSchemaBrowser:
		cv.Browser, cmd = cv.Browser.Update(msg)
	case PaneEditor:
		cv.Editor, cmd = cv.Editor.Update(msg)
	case PaneResults:
		cv.Results, cmd = cv.Results.Update(msg)
	}

	return cv, cmd
}

// View renders the view
func (cv *ConnectedView) View(width, height int) string {
	cv.width = width
	cv.height = height
	cv.updateDimensions()

	// Calculate layout
	leftPanelWidth := (width * cv.LeftWidth) / 100
	rightPanelWidth := width - leftPanelWidth

	// Right panel split vertically (50/50)
	editorHeight := (height - 2) / 2  // -2 for status bar
	resultsHeight := height - editorHeight - 2

	// Render components
	browserView := cv.Browser.View()
	editorView := cv.Editor.View()
	resultsView := cv.Results.View()
	statusView := cv.StatusBar.View()

	// Apply focus styles
	browserStyle := lipgloss.NewStyle().Width(leftPanelWidth).Height(height - 2)
	editorStyle := lipgloss.NewStyle().Width(rightPanelWidth).Height(editorHeight)
	resultsStyle := lipgloss.NewStyle().Width(rightPanelWidth).Height(resultsHeight)

	if cv.FocusedPane == PaneSchemaBrowser {
		browserStyle = browserStyle.BorderForeground(lipgloss.Color("63"))
	}
	if cv.FocusedPane == PaneEditor {
		editorStyle = editorStyle.BorderForeground(lipgloss.Color("63"))
	}
	if cv.FocusedPane == PaneResults {
		resultsStyle = resultsStyle.BorderForeground(lipgloss.Color("63"))
	}

	// Layout: left panel | right panel (editor + results)
	rightPanel := lipgloss.JoinVertical(
		lipgloss.Top,
		editorStyle.Render(editorView),
		resultsStyle.Render(resultsView),
	)

	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		browserStyle.Render(browserView),
		rightPanel,
	)

	// Add status bar at bottom
	fullView := lipgloss.JoinVertical(
		lipgloss.Top,
		mainContent,
		statusView,
	)

	return fullView
}

// updateDimensions updates component dimensions
func (cv *ConnectedView) updateDimensions() {
	leftPanelWidth := (cv.width * cv.LeftWidth) / 100
	rightPanelWidth := cv.width - leftPanelWidth

	editorHeight := (cv.height - 2) / 2
	resultsHeight := cv.height - editorHeight - 2

	cv.Browser.SetDimensions(leftPanelWidth, cv.height-2)
	cv.Editor.SetDimensions(rightPanelWidth, editorHeight)
	cv.Results.SetDimensions(rightPanelWidth, resultsHeight)
	cv.StatusBar.SetWidth(cv.width)
}
