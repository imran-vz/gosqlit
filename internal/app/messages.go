package app

import (
	"time"

	"github.com/imran-vz/gosqlit/internal/db"
	"github.com/imran-vz/gosqlit/internal/ui/modal"
)

// Connection lifecycle
type ConnectRequestMsg struct {
	Config SavedConnection
}

type ConnectSuccessMsg struct {
	ConnID     string
	Connection db.Connection
}

type ConnectErrorMsg struct {
	ConnID string
	Err    error
}

type DisconnectMsg struct {
	ConnID string
}

// Database operations
type LoadSchemasMsg struct {
	ConnID string
}

type SchemasLoadedMsg struct {
	ConnID  string
	Schemas []db.Schema
	Err     error
}

type ExecuteQueryMsg struct {
	ConnID string
	SQL    string
	Offset int
}

type QueryResultMsg struct {
	ConnID  string
	Result  db.QueryResult
	Err     error
	Elapsed time.Duration
}

type QueryCancelMsg struct {
	ConnID string
}

type LoadMoreResultsMsg struct {
	ConnID string
}

type ExportCSVMsg struct {
	ConnID    string
	FilePath  string
	ExportAll bool
}

// UI interactions
type TableSelectedMsg struct {
	Schema string
	Table  string
}

type OpenModalMsg struct {
	Modal modal.Modal
}

type CloseModalMsg struct{}

type RefreshSchemasMsg struct {
	ConnID string
}

// Tab management
type NewTabMsg struct {
	ConnID string
}

type CloseTabMsg struct {
	Index int
}

type SwitchTabMsg struct {
	Index int
}

// SavedConnection from config
type SavedConnection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}
