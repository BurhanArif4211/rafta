package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    local_ip TEXT NOT NULL,
    hostname TEXT,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS note_folders (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES note_folders(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notes (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT,
    folder_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (folder_id) REFERENCES note_folders(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS todo_folders (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES todo_folders(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS todos (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    folder_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (folder_id) REFERENCES todo_folders(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS todo_steps (
    id TEXT PRIMARY KEY,
    todo_id TEXT NOT NULL,
    description TEXT NOT NULL,
    completed BOOLEAN DEFAULT 0,
    display_order INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (todo_id) REFERENCES todos(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_note_folders_parent ON note_folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_notes_folder ON notes(folder_id);
CREATE INDEX IF NOT EXISTS idx_todo_folders_parent ON todo_folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_todos_folder ON todos(folder_id);
CREATE INDEX IF NOT EXISTS idx_todo_steps_todo ON todo_steps(todo_id);
`

// InitDB opens the database, applies pragmas, and creates tables.
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys and set other pragmas
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;", // better concurrency
		"PRAGMA synchronous = NORMAL;",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			log.Printf("warning: failed to set pragma %s: %v", pragma, err)
		}
	}

	// Create tables
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return db, nil
}

func GetDatabasePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to temporary directory (should never happen on Android)
		return filepath.Join(os.TempDir(), "rafta-main.db")
	}
	// Create a subdirectory for your app to avoid cluttering
	// appDir := filepath.Join(configDir, "rafta")
	// os.MkdirAll(appDir, 0700) // ensure directory exists
	return filepath.Join(configDir, "rafta-main.db")
}
