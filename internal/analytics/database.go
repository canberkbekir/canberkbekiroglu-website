package analytics

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps an SQLite database for analytics tracking.
type DB struct {
	conn *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS visitors (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	ip_address TEXT    NOT NULL UNIQUE,
	city       TEXT    NOT NULL DEFAULT 'Unknown',
	created_at DATETIME NOT NULL DEFAULT (datetime('now')),
	updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS page_views (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	visitor_id INTEGER NOT NULL,
	page_path  TEXT    NOT NULL,
	view_count INTEGER NOT NULL DEFAULT 1,
	date       TEXT    NOT NULL DEFAULT (date('now')),
	FOREIGN KEY (visitor_id) REFERENCES visitors(id) ON DELETE CASCADE,
	UNIQUE(visitor_id, page_path, date)
);

CREATE TABLE IF NOT EXISTS request_metrics (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	path        TEXT    NOT NULL,
	method      TEXT    NOT NULL,
	status_code INTEGER NOT NULL,
	duration_ms INTEGER NOT NULL,
	date        TEXT    NOT NULL DEFAULT (date('now')),
	hour        INTEGER NOT NULL DEFAULT (CAST(strftime('%H', 'now') AS INTEGER))
);

CREATE INDEX IF NOT EXISTS idx_request_metrics_date ON request_metrics(date);
CREATE INDEX IF NOT EXISTS idx_visitors_created ON visitors(created_at);
CREATE INDEX IF NOT EXISTS idx_page_views_date ON page_views(date);
`

// NewDB opens (or creates) the SQLite database at the given path and applies the schema.
func NewDB(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", path+"?_busy_timeout=5000")
	if err != nil {
		return nil, err
	}

	// SQLite only supports one writer at a time. Limiting to a single open
	// connection serialises writes from concurrent goroutines and prevents
	// "database is locked" errors.
	conn.SetMaxOpenConns(1)

	// Enable WAL mode and foreign keys.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}

	log.Printf("Analytics database ready at %s", path)
	return &DB{conn: conn}, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}
