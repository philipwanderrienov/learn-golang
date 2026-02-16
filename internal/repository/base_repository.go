package repository

import (
	"context"
	"database/sql"
)

// Repository is the interface that all repositories must implement.
// Similar to .NET's IRepository<T>, but tailored for Go's context-first approach.
type Repository interface {
	// ScanRow executes a SELECT query that returns a single row.
	// The scanFn callback handles reading the row into a destination.
	ScanRow(ctx context.Context, query string, scanFn func(*sql.Row) error, args ...interface{}) error

	// ScanRows executes a SELECT query that returns multiple rows.
	// The scanFn callback iterates through rows and returns any errors.
	ScanRows(ctx context.Context, query string, scanFn func(*sql.Rows) error, args ...interface{}) error

	// ExecUpdate executes an INSERT, UPDATE, or DELETE query.
	// Returns error if the operation fails.
	ExecUpdate(ctx context.Context, query string, args ...interface{}) error
}

// BaseRepository is a generic repository implementation that all domain repositories can embed.
// It provides common database operations (like .NET's Repository<T> base class).
type BaseRepository struct {
	db *sql.DB
}

// NewBaseRepository creates a new base repository with a database connection pool.
func NewBaseRepository(db *sql.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// ScanRow executes a SELECT query and scans a single row using the provided scanFn.
// This avoids repeating r.db.QueryRowContext(...).Scan(...) boilerplate in each repo.
// The scanFn callback is responsible for reading the row data.
func (br *BaseRepository) ScanRow(ctx context.Context, query string, scanFn func(*sql.Row) error, args ...interface{}) error {
	return scanFn(br.db.QueryRowContext(ctx, query, args...))
}

// ExecUpdate executes an INSERT, UPDATE, or DELETE query.
// This avoids repeating r.db.ExecContext(...) boilerplate in each repo.
func (br *BaseRepository) ExecUpdate(ctx context.Context, query string, args ...interface{}) error {
	_, err := br.db.ExecContext(ctx, query, args...)
	return err
}

// ScanRows executes a SELECT query that returns multiple rows and iterates using the provided scanFn.
// This avoids repeating QueryContext + defer Close boilerplate.
// The scanFn callback is responsible for iterating rows.Next() and scanning each row.
func (br *BaseRepository) ScanRows(ctx context.Context, query string, scanFn func(*sql.Rows) error, args ...interface{}) error {
	rows, err := br.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanFn(rows)
}
