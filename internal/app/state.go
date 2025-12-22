package app

import (
	"github.com/imran-vz/gosqlit/internal/db"
	"github.com/imran-vz/gosqlit/internal/ui/connected"
	"github.com/imran-vz/gosqlit/internal/ui/modal"
)

// ViewType represents current view
type ViewType int

const (
	ViewExplorer ViewType = iota
	ViewConnected
)

// App is root model
type App struct {
	// Core state
	masterPassword string
	connections    *ConnectionManager
	tabs           []Tab
	currentTabIdx  int
	currentView    ViewType
	configMgr      interface{} // config.Manager - for saving connections

	// UI state
	windowWidth  int
	windowHeight int
	activeModal  modal.Modal

	// Views
	explorerView interface{} // Will be ui/explorer.ExplorerView
}

// ConnectionManager manages active connections
type ConnectionManager struct {
	saved  []SavedConnection
	active map[string]db.Connection // connID → connection
}

// NewConnectionManager creates manager
func NewConnectionManager(saved []SavedConnection) *ConnectionManager {
	return &ConnectionManager{
		saved:  saved,
		active: make(map[string]db.Connection),
	}
}

// GetConnection retrieves active connection
func (cm *ConnectionManager) GetConnection(id string) (db.Connection, bool) {
	conn, ok := cm.active[id]
	return conn, ok
}

// AddConnection adds active connection
func (cm *ConnectionManager) AddConnection(id string, conn db.Connection) {
	cm.active[id] = conn
}

// RemoveConnection removes connection
func (cm *ConnectionManager) RemoveConnection(id string) {
	if conn, ok := cm.active[id]; ok {
		conn.Close()
		delete(cm.active, id)
	}
}

// Tab represents connection tab
type Tab struct {
	ID     string
	ConnID string
	View   *connected.ConnectedView
}
