package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/imran-vz/gosqlit/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connection implements db.Connection for PostgreSQL
type Connection struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

// Query executes SQL query with pagination
func (c *Connection) Query(ctx context.Context, sql string, limit int, offset int) (db.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// TODO: Implement pagination with proper query breakup
	// Add LIMIT and OFFSET to query
	// paginatedSQL := fmt.Sprintf("%s", sql, limit, offset)

	rows, err := c.pool.Query(ctx, sql)
	if err != nil {
		return db.QueryResult{}, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	fieldDescs := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescs))
	for i, field := range fieldDescs {
		columns[i] = string(field.Name)
	}

	// Collect rows
	var resultRows [][]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return db.QueryResult{}, fmt.Errorf("failed to scan row: %w", err)
		}
		resultRows = append(resultRows, values)
	}

	if err := rows.Err(); err != nil {
		return db.QueryResult{}, fmt.Errorf("rows error: %w", err)
	}

	// Check if there are more rows
	hasMore := len(resultRows) == limit

	return db.QueryResult{
		Columns:    columns,
		Rows:       resultRows,
		RowCount:   len(resultRows),
		HasMore:    hasMore,
		ResultSets: []db.QueryResultSet{}, // Single result set for now
	}, nil
}

// ListSchemas returns all schemas
func (c *Connection) ListSchemas(ctx context.Context) ([]db.Schema, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	query := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		ORDER BY schema_name
	`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()

	var schemas []db.Schema
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}

		// Get tables for this schema
		tables, err := c.ListTables(ctx, schemaName)
		if err != nil {
			return nil, fmt.Errorf("failed to list tables for schema %s: %w", schemaName, err)
		}

		schemas = append(schemas, db.Schema{
			Name:   schemaName,
			Tables: tables,
		})
	}

	return schemas, nil
}

// ListTables returns tables in a schema
func (c *Connection) ListTables(ctx context.Context, schema string) ([]db.Table, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := c.pool.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []db.Table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}

		tables = append(tables, db.Table{
			Name:   tableName,
			Schema: schema,
		})
	}

	return tables, nil
}

// GetTableInfo returns table metadata
func (c *Connection) GetTableInfo(ctx context.Context, schema, table string) (db.TableInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	query := `
		SELECT
			column_name,
			data_type,
			is_nullable,
			COALESCE(
				(SELECT constraint_type
				 FROM information_schema.table_constraints tc
				 JOIN information_schema.key_column_usage kcu
				   ON tc.constraint_name = kcu.constraint_name
				  AND tc.table_schema = kcu.table_schema
				 WHERE tc.table_schema = $1
				   AND tc.table_name = $2
				   AND kcu.column_name = c.column_name
				 LIMIT 1),
				''
			) as key_type
		FROM information_schema.columns c
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`

	rows, err := c.pool.Query(ctx, query, schema, table)
	if err != nil {
		return db.TableInfo{}, fmt.Errorf("failed to get table info: %w", err)
	}
	defer rows.Close()

	var columns []db.ColumnInfo
	for rows.Next() {
		var col db.ColumnInfo
		var nullable string
		var keyType string

		if err := rows.Scan(&col.Name, &col.Type, &nullable, &keyType); err != nil {
			return db.TableInfo{}, fmt.Errorf("failed to scan column: %w", err)
		}

		col.Nullable = (nullable == "YES")
		col.Key = convertKeyType(keyType)

		columns = append(columns, col)
	}

	return db.TableInfo{
		Columns: columns,
	}, nil
}

// Ping checks if connection is alive
func (c *Connection) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.pool.Ping(ctx)
}

// Close closes the connection
func (c *Connection) Close() error {
	c.pool.Close()
	return nil
}

// SetTimeout sets query timeout
func (c *Connection) SetTimeout(duration time.Duration) {
	c.timeout = duration
}

// convertKeyType converts PostgreSQL constraint type to simplified key indicator
func convertKeyType(constraintType string) string {
	switch constraintType {
	case "PRIMARY KEY":
		return "PRI"
	case "UNIQUE":
		return "UNI"
	case "FOREIGN KEY":
		return "MUL"
	default:
		return ""
	}
}
