package connected

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/imran-vz/gosqlit/internal/debug"
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
	DebugMode    bool // debug flag

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
	debug.Logf("Creating connected view for connID: %s", connID)
	
	return &ConnectedView{
		ConnID:      connID,
		FocusedPane: PaneEditor,
		LeftWidth:   25, // 25% for left panel
		DebugMode:   debug.Enabled(),

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
		debug.LogKey(keyMsg.String(), "connected_view")
		
		switch keyMsg.String() {
		case "tab":
			// Cycle focus
			debug.Logf("Tab pressed - current pane: %d", cv.FocusedPane)
			switch cv.FocusedPane {
			case PaneSchemaBrowser:
				cv.FocusedPane = PaneEditor
			case PaneEditor:
				cv.FocusedPane = PaneResults
			case PaneResults:
				cv.FocusedPane = PaneSchemaBrowser
			}
			debug.Logf("Tab handled - new pane: %d", cv.FocusedPane)
			return cv, nil
		}
	}

	// Delegate to focused pane
	switch cv.FocusedPane {
	case PaneSchemaBrowser:
		debug.Logf("Delegating to schema browser")
		cv.Browser, cmd = cv.Browser.Update(msg)
	case PaneEditor:
		debug.Logf("Delegating to query editor (pane focused)")
		cv.Editor, cmd = cv.Editor.Update(msg)
	case PaneResults:
		debug.Logf("Delegating to results table")
		cv.Results, cmd = cv.Results.Update(msg)
	}

	return cv, cmd
}

// View renders the view
func (cv *ConnectedView) View(width, height int) string {
	cv.width = width
	cv.height = height
	cv.updateDimensions()

	// Border takes 2 chars each direction (top+bottom, left+right)
	const borderSize = 2

	// Calculate available height (account for status bar and debug overlay)
	availableHeight := height - 1 // status bar
	if cv.DebugMode {
		availableHeight -= 1 // Debug overlay takes 1 line
	}

	// Calculate layout - account for borders
	leftPanelWidth := (width * cv.LeftWidth) / 100
	rightPanelWidth := width - leftPanelWidth

	// Content dimensions (subtract border from total)
	browserContentWidth := leftPanelWidth - borderSize
	browserContentHeight := availableHeight - borderSize

	// Right panel heights - split available height between editor and results
	rightPanelHeight := availableHeight
	editorContentHeight := (rightPanelHeight - borderSize*2) / 2 // two borders for two panels
	resultsContentHeight := rightPanelHeight - borderSize*2 - editorContentHeight

	editorContentWidth := rightPanelWidth - borderSize
	resultsContentWidth := rightPanelWidth - borderSize

	// Render components
	browserView := cv.Browser.View()
	editorView := cv.Editor.View()
	resultsView := cv.Results.View()
	statusView := cv.StatusBar.View()

	// Apply focus styles with borders
	defaultBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	browserStyle := defaultBorderStyle.Width(browserContentWidth).Height(browserContentHeight)
	editorStyle := defaultBorderStyle.Width(editorContentWidth).Height(editorContentHeight)
	resultsStyle := defaultBorderStyle.Width(resultsContentWidth).Height(resultsContentHeight)

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

	// Add debug overlay if enabled
	if cv.DebugMode {
		debugStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("90")).
			Bold(true).
			Padding(0, 1).
			MaxWidth(cv.width)

		debugInfo := fmt.Sprintf("DEBUG: FocusedPane=%d | Dimensions=%dx%d | LeftWidth=%d%%",
			cv.FocusedPane, cv.width, cv.height, cv.LeftWidth)

		debugOverlay := debugStyle.Render(debugInfo)

		// Add to top of view
		fullView = lipgloss.JoinVertical(
			lipgloss.Top,
			debugOverlay,
			fullView,
		)
	}

	return fullView
}

// updateDimensions updates component dimensions
func (cv *ConnectedView) updateDimensions() {
	const borderSize = 2

	// Calculate available height (account for status bar and debug overlay)
	availableHeight := cv.height - 1 // status bar
	if cv.DebugMode {
		availableHeight -= 1 // Debug overlay takes 1 line
	}

	leftPanelWidth := (cv.width * cv.LeftWidth) / 100
	rightPanelWidth := cv.width - leftPanelWidth

	// Content dimensions (subtract border)
	browserContentWidth := leftPanelWidth - borderSize
	browserContentHeight := availableHeight - borderSize

	rightPanelHeight := availableHeight
	editorContentHeight := (rightPanelHeight - borderSize*2) / 2
	resultsContentHeight := rightPanelHeight - borderSize*2 - editorContentHeight

	editorContentWidth := rightPanelWidth - borderSize
	resultsContentWidth := rightPanelWidth - borderSize

	cv.Browser.SetDimensions(browserContentWidth, browserContentHeight)
	cv.Editor.SetDimensions(editorContentWidth, editorContentHeight)
	cv.Results.SetDimensions(resultsContentWidth, resultsContentHeight)
	cv.StatusBar.SetWidth(cv.width)
}
