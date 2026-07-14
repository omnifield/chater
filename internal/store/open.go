package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (no cgo), registers "sqlite"
)

// Open returns a *sql.DB for the given SQLite path with foreign-key enforcement
// and a busy timeout enabled. Pass ":memory:" for an ephemeral database.
//
// The driver name is the only SQLite-specific detail here; switching to Postgres
// is a driver + DSN change (canon: DB choice is config, not code branches).
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}
	return db, nil
}
