package db

import (
	"context"
	"time"
)

// Driver interface for database drivers
type Driver interface {
	Name() string
	Connect(ctx context.Context, config ConnConfig) (Connection, error)
	DefaultPort() int
	RequiredFields() []string
}

// Connection interface for active database connections
type Connection interface {
	Query(ctx context.Context, sql string, limit int, offset int) (QueryResult, error)
	ListSchemas(ctx context.Context) ([]Schema, error)
	ListTables(ctx context.Context, schema string) ([]Table, error)
	GetTableInfo(ctx context.Context, schema, table string) (TableInfo, error)
	Ping(ctx context.Context) error
	Close() error
	SetTimeout(duration time.Duration)
}

// ConnConfig holds connection configuration
type ConnConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// QueryResult holds query results
type QueryResult struct {
	Columns    []string
	Rows       [][]interface{}
	RowCount   int
	HasMore    bool
	ResultSets []QueryResultSet
}

// QueryResultSet for multiple result sets
type QueryResultSet struct {
	Columns []string
	Rows    [][]interface{}
}

// Schema represents database schema
type Schema struct {
	Name   string
	Tables []Table
}

// Table represents database table
type Table struct {
	Name   string
	Schema string
}

// TableInfo holds table metadata
type TableInfo struct {
	Columns []ColumnInfo
}

// ColumnInfo holds column metadata
type ColumnInfo struct {
	Name     string
	Type     string
	Nullable bool
	Key      string // PRI, UNI, MUL, ""
}
