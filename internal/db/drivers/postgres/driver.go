package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/imran-vz/gosqlit/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Driver implements db.Driver for PostgreSQL
type Driver struct{}

// Name returns driver name
func (d *Driver) Name() string {
	return "postgres"
}

// DefaultPort returns default PostgreSQL port
func (d *Driver) DefaultPort() int {
	return 5432
}

// RequiredFields returns required connection fields
func (d *Driver) RequiredFields() []string {
	return []string{"host", "port", "username", "password", "database"}
}

// Connect establishes connection to PostgreSQL
func (d *Driver) Connect(ctx context.Context, config db.ConnConfig) (db.Connection, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=prefer",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Set pool settings
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Connection{
		pool:    pool,
		timeout: 30 * time.Second, // Default timeout
	}, nil
}
