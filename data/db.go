package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

// DBEngine wraps the standard database/sql connection pool and provides
// integrated transaction management and query building capabilities.
type DBEngine struct {
	SQL     *sql.DB
	Builder squirrel.StatementBuilderType
	Dialect string
}

// NewDBEngine initializes a new database connection pool.
func NewDBEngine(driverName, dataSourceName string) (*DBEngine, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("ztatic/data: failed to open db: %w", err)
	}
	
	// Ensure the connection is actually valid
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ztatic/data: failed to ping db: %w", err)
	}

	var builder squirrel.StatementBuilderType
	driverLower := strings.ToLower(driverName)
	if strings.Contains(driverLower, "postgres") || strings.Contains(driverLower, "pgx") {
		builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).RunWith(db)
	} else {
		builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question).RunWith(db)
	}

	return &DBEngine{
		SQL:     db,
		Builder: builder,
		Dialect: driverName,
	}, nil
}

// Close gracefully terminates the database connection pool.
func (e *DBEngine) Close() error {
	return e.SQL.Close()
}

// Transaction executes a closure within an ACID-compliant database transaction.
// It automatically handles Commit on success, and Rollback on error or panic.
func (e *DBEngine) Transaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := e.SQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Ensure rollback on panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // Re-throw the panic after rolling back
		}
	}()

	// Execute the business logic
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	// Commit if successful
	return tx.Commit()
}
