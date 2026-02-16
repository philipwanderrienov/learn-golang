package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// ConnectDB opens a sql.DB using the provided Postgres connection string.
// Caller should call Close() on the returned *sql.DB when finished.
func ConnectDB(connStr string) (*sql.DB, error) {
	// database/sql manages a pool for us; lib/pq is the driver.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Ping to verify connectivity and credentials early.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
