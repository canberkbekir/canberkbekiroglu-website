package admin

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps an SQLite database for admin/project management.
type DB struct {
	conn *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS sections (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	name       TEXT NOT NULL,
	slug       TEXT NOT NULL UNIQUE,
	sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS projects (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	title        TEXT NOT NULL,
	slug         TEXT NOT NULL UNIQUE,
	year         INTEGER NOT NULL,
	description  TEXT NOT NULL DEFAULT '',
	image        TEXT NOT NULL DEFAULT '',
	hover_image  TEXT NOT NULL DEFAULT '',
	category     TEXT NOT NULL DEFAULT '',
	company      TEXT NOT NULL DEFAULT '',
	role         TEXT NOT NULL DEFAULT '',
	external_url TEXT NOT NULL DEFAULT '',
	highlighted  INTEGER NOT NULL DEFAULT 0,
	sort_order   INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS project_tags (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id INTEGER NOT NULL,
	tag        TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS filter_tags (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	slug       TEXT NOT NULL UNIQUE,
	label      TEXT NOT NULL,
	is_gold    INTEGER NOT NULL DEFAULT 0,
	sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS project_sections (
	project_id INTEGER NOT NULL,
	section_id INTEGER NOT NULL,
	PRIMARY KEY (project_id, section_id),
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
	FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS project_filter_tags (
	project_id    INTEGER NOT NULL,
	filter_tag_id INTEGER NOT NULL,
	PRIMARY KEY (project_id, filter_tag_id),
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
	FOREIGN KEY (filter_tag_id) REFERENCES filter_tags(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sessions (
	token      TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL DEFAULT (datetime('now')),
	expires_at DATETIME NOT NULL
);
`

// NewDB opens (or creates) the admin SQLite database at the given path.
func NewDB(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", path+"?_busy_timeout=5000")
	if err != nil {
		return nil, err
	}

	if _, err := conn.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}

	log.Printf("Admin database ready at %s", path)
	return &DB{conn: conn}, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}
