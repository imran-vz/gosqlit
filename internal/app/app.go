package app

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/imran-vz/gosqlit/internal/config"
	"github.com/imran-vz/gosqlit/internal/db"
	"github.com/imran-vz/gosqlit/internal/ui/connected"
	"github.com/imran-vz/gosqlit/internal/ui/explorer"
	"github.com/imran-vz/gosqlit/internal/ui/modal"
)

// New creates new app
func New(configMgr *config.Manager) *App {
	connections := configMgr.GetConnections()

	return &App{
		connections:  NewConnectionManager(ToSavedConnections(connections)),
		tabs:         []Tab{},
		currentView:  ViewExplorer,
		explorerView: explorer.NewExplorer(connections),
		configMgr:    configMgr,
	}
}

// Init initializes the app
func (a *App) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.windowWidth = msg.Width
		a.windowHeight = msg.Height

		// Pass to current view
		if a.currentView == ViewExplorer && a.explorerView != nil {
			exp := a.explorerView.(*explorer.ExplorerView)
			exp, _ = exp.Update(msg)
			a.explorerView = exp
		}

	case tea.KeyMsg:
		// Handle modal first
		if a.activeModal != nil {
			model, cmd := a.activeModal.Update(msg)
			if newModal, ok := model.(modal.Modal); ok {
				a.activeModal = newModal

				// Check if modal closed
				if !newModal.IsOpen() {
					// Check if it's a connection form that was submitted
					if connForm, ok := newModal.(*modal.ConnectionFormModal); ok {
						if connForm.IsSubmitted() {
							// Save connection
							conn := connForm.GetConnection()

							// Save to config file
							if cfgMgr, ok := a.configMgr.(*config.Manager); ok {
								if err := cfgMgr.AddConnection(conn); err != nil {
									fmt.Printf("Failed to save connection: %v\n", err)
								}
							}

							// Add to in-memory list
							a.connections.saved = append(a.connections.saved, SavedConnection{
								ID:       conn.ID,
								Name:     conn.Name,
								Driver:   conn.Driver,
								Host:     conn.Host,
								Port:     conn.Port,
								Username: conn.Username,
								Password: conn.Password,
								Database: conn.Database,
							})

							// Update explorer view
							if exp, ok := a.explorerView.(*explorer.ExplorerView); ok {
								// Get fresh connections from config
								if cfgMgr, ok := a.configMgr.(*config.Manager); ok {
									exp.SetConnections(cfgMgr.GetConnections())
									a.explorerView = exp
								}
							}
						}
					}
					a.activeModal = nil
				}
			}

			return a, cmd
		}

		// Global shortcuts
		switch msg.String() {
		case "ctrl+c", "q":
			if a.currentView == ViewExplorer {
				return a, tea.Quit
			}
		}

		// Delegate to current view
		return a.updateCurrentView(msg)

	// Handle connection messages
	case ConnectRequestMsg:
		return a, a.connectCmd(msg)

	case ConnectSuccessMsg:
		// Create new tab with connected view
		connInfo := fmt.Sprintf("%s @ %s", msg.ConnID, "database")
		tab := Tab{
			ID:     msg.ConnID,
			ConnID: msg.ConnID,
			View:   connected.NewConnectedView(msg.ConnID, connInfo),
		}
		a.tabs = append(a.tabs, tab)
		a.currentTabIdx = len(a.tabs) - 1
		a.currentView = ViewConnected

		// Load schemas
		return a, a.loadSchemasCmd(msg.ConnID)

	case ConnectErrorMsg:
		// Show error and go back to explorer
		fmt.Printf("Connection error: %v\n", msg.Err)
		a.currentView = ViewExplorer
		return a, nil

	case SchemasLoadedMsg:
		// Update schema browser
		if a.currentView == ViewConnected && a.currentTabIdx >= 0 && a.currentTabIdx < len(a.tabs) {
			tab := &a.tabs[a.currentTabIdx]
			if tab.ConnID == msg.ConnID {
				if msg.Err != nil {
					fmt.Printf("Failed to load schemas: %v\n", msg.Err)
				} else {
					tab.View.Browser.SetSchemas(msg.Schemas)
				}
			}
		}

	case ExecuteQueryMsg:
		return a, a.executeQueryCmd(msg)

	case QueryResultMsg:
		// Update results table
		if a.currentView == ViewConnected && a.currentTabIdx >= 0 && a.currentTabIdx < len(a.tabs) {
			tab := &a.tabs[a.currentTabIdx]
			if tab.ConnID == msg.ConnID {
				if msg.Err != nil {
					tab.View.StatusBar.SetError(msg.Err.Error())
				} else {
					tab.View.Results.SetData(msg.Result)
					tab.View.StatusBar.SetQueryResult(msg.Result.RowCount, msg.Elapsed)
				}
				tab.View.QueryRunning = false
			}
		}
	}

	return a, nil
}

// View renders the app
func (a *App) View() string {
	// Show modal if active
	if a.activeModal != nil {
		return a.activeModal.ViewSized(a.windowWidth, a.windowHeight)
	}

	// Show current view
	switch a.currentView {
	case ViewExplorer:
		if a.explorerView != nil {
			return a.explorerView.(*explorer.ExplorerView).View()
		}
		return "Loading..."

	case ViewConnected:
		if a.currentTabIdx >= 0 && a.currentTabIdx < len(a.tabs) {
			tab := a.tabs[a.currentTabIdx]
			return tab.View.View(a.windowWidth, a.windowHeight)
		}
		return "No active connection"

	default:
		return "Unknown view"
	}
}

// updateCurrentView delegates message to active view
func (a *App) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch a.currentView {
	case ViewExplorer:
		return a.updateExplorer(msg)
	case ViewConnected:
		return a.updateConnected(msg)
	}
	return a, nil
}

// updateExplorer handles explorer view messages
func (a *App) updateExplorer(msg tea.Msg) (tea.Model, tea.Cmd) {
	exp := a.explorerView.(*explorer.ExplorerView)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			// Connect to selected connection
			if conn := exp.GetSelectedConnection(); conn != nil {
				a.currentView = ViewConnected
				return a, func() tea.Msg {
					return ConnectRequestMsg{
						Config: SavedConnection{
							ID:       conn.ID,
							Name:     conn.Name,
							Driver:   conn.Driver,
							Host:     conn.Host,
							Port:     conn.Port,
							Username: conn.Username,
							Password: conn.Password,
							Database: conn.Database,
						},
					}
				}
			}
		case "n":
			// Open new connection modal
			a.activeModal = modal.NewConnectionForm(nil)
			return a, nil
		case "d":
			// Delete selected connection
			if conn := exp.GetSelectedConnection(); conn != nil {
				// Delete from config file
				if cfgMgr, ok := a.configMgr.(*config.Manager); ok {
					if err := cfgMgr.DeleteConnection(conn.ID); err != nil {
						fmt.Printf("Failed to delete connection: %v\n", err)
					}
				}

				// Remove from in-memory list
				filtered := []SavedConnection{}
				for _, c := range a.connections.saved {
					if c.ID != conn.ID {
						filtered = append(filtered, c)
					}
				}
				a.connections.saved = filtered

				// Update explorer view
				if cfgMgr, ok := a.configMgr.(*config.Manager); ok {
					exp.SetConnections(cfgMgr.GetConnections())
					a.explorerView = exp
				}
			}
			return a, nil
		case "e":
			// Edit selected connection
			if conn := exp.GetSelectedConnection(); conn != nil {
				a.activeModal = modal.NewConnectionForm(conn)
			}
			return a, nil
		}
	}

	newExp, cmd := exp.Update(msg)
	a.explorerView = newExp
	return a, cmd
}

// updateConnected handles connected view messages
func (a *App) updateConnected(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.currentTabIdx < 0 || a.currentTabIdx >= len(a.tabs) {
		return a, nil
	}

	tab := &a.tabs[a.currentTabIdx]

	// Handle global shortcuts first
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+w":
			// Close current tab
			a.tabs = append(a.tabs[:a.currentTabIdx], a.tabs[a.currentTabIdx+1:]...)
			if len(a.tabs) == 0 {
				a.currentView = ViewExplorer
			} else if a.currentTabIdx >= len(a.tabs) {
				a.currentTabIdx = len(a.tabs) - 1
			}
			return a, nil

		case "ctrl+t":
			// New tab - switch to explorer
			a.currentView = ViewExplorer
			return a, nil

		case "ctrl+enter":
			// Execute query
			sql := tab.View.Editor.GetContent()
			if sql != "" {
				tab.View.QueryRunning = true
				tab.View.StatusBar.SetQueryRunning(true)
				return a, func() tea.Msg {
					return ExecuteQueryMsg{
						ConnID: tab.ConnID,
						SQL:    sql,
						Offset: 0,
					}
				}
			}
			return a, nil

		case "ctrl+k":
			// Cancel query
			if tab.View.QueryRunning && tab.View.CancelFunc != nil {
				tab.View.CancelFunc()
				tab.View.QueryRunning = false
				tab.View.StatusBar.SetError("Query cancelled")
			}
			return a, nil

		case "f5":
			// Refresh schemas
			return a, a.loadSchemasCmd(tab.ConnID)
		}

		// Handle table selection in schema browser
		if tab.View.FocusedPane == connected.PaneSchemaBrowser {
			if keyMsg.String() == "enter" {
				if schema, table, ok := tab.View.Browser.GetSelectedTable(); ok {
					// Auto-generate SELECT query
					sql := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 100", schema, table)
					tab.View.Editor.SetContent(sql)
					tab.View.FocusedPane = connected.PaneEditor
					return a, nil
				}
			}
		}
	}

	// Update tab view
	newView, cmd := tab.View.Update(msg)
	tab.View = newView
	return a, cmd
}

// connectCmd initiates connection
func (a *App) connectCmd(msg ConnectRequestMsg) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get driver
		driver, err := db.GetDriver(msg.Config.Driver)
		if err != nil {
			return ConnectErrorMsg{
				ConnID: msg.Config.ID,
				Err:    fmt.Errorf("driver not found: %w", err),
			}
		}

		// Connect
		conn, err := driver.Connect(ctx, db.ConnConfig{
			Host:     msg.Config.Host,
			Port:     msg.Config.Port,
			Username: msg.Config.Username,
			Password: msg.Config.Password,
			Database: msg.Config.Database,
		})
		if err != nil {
			return ConnectErrorMsg{
				ConnID: msg.Config.ID,
				Err:    fmt.Errorf("connection failed: %w", err),
			}
		}

		// Store connection
		a.connections.AddConnection(msg.Config.ID, conn)

		return ConnectSuccessMsg{
			ConnID:     msg.Config.ID,
			Connection: conn,
		}
	}
}

// loadSchemasCmd loads schemas for a connection
func (a *App) loadSchemasCmd(connID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		conn, ok := a.connections.GetConnection(connID)
		if !ok {
			return SchemasLoadedMsg{
				ConnID: connID,
				Err:    fmt.Errorf("connection not found"),
			}
		}

		schemas, err := conn.ListSchemas(ctx)
		return SchemasLoadedMsg{
			ConnID:  connID,
			Schemas: schemas,
			Err:     err,
		}
	}
}

// executeQueryCmd executes SQL query with cancellation support
func (a *App) executeQueryCmd(msg ExecuteQueryMsg) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // Always call cancel to prevent context leak

		// Store cancel func in tab for Ctrl+K cancellation
		if a.currentTabIdx >= 0 && a.currentTabIdx < len(a.tabs) {
			a.tabs[a.currentTabIdx].View.CancelFunc = cancel
		}

		conn, ok := a.connections.GetConnection(msg.ConnID)
		if !ok {
			return QueryResultMsg{
				ConnID: msg.ConnID,
				Err:    fmt.Errorf("connection not found"),
			}
		}

		result, err := conn.Query(ctx, msg.SQL, 1000, msg.Offset)
		return QueryResultMsg{
			ConnID:  msg.ConnID,
			Result:  result,
			Err:     err,
			Elapsed: time.Since(start),
		}
	}
}

// ToSavedConnections converts config connections to app connections
func ToSavedConnections(configs []config.SavedConnection) []SavedConnection {
	result := make([]SavedConnection, len(configs))
	for i, c := range configs {
		result[i] = SavedConnection{
			ID:       c.ID,
			Name:     c.Name,
			Driver:   c.Driver,
			Host:     c.Host,
			Port:     c.Port,
			Username: c.Username,
			Password: c.Password,
			Database: c.Database,
		}
	}
	return result
}
