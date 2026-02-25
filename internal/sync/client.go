package sync

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client fetches data from a remote device and replaces the local database.
type Client struct {
	db *sql.DB
}

func NewClient(db *sql.DB) *Client {
	return &Client{db: db}
}

// Pull fetches sync data from the given IP:port and replaces local tables.
func (c *Client) Pull(serverAddr string) error {

	serverAddr = fmt.Sprintf("%s:4211", serverAddr)
	url := fmt.Sprintf("http://%s/api/sync", serverAddr)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	var data SyncData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Replace local database in a transaction
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Disable foreign keys temporarily to allow deletion in any order
	if _, err := tx.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return err
	}
	// Re-enable at the end
	defer tx.Exec("PRAGMA foreign_keys = ON")

	// Clear existing data (order doesn't matter with FK off)
	tables := []string{"todo_steps", "todos", "todo_folders", "notes", "note_folders"}
	for _, table := range tables {
		if _, err := tx.Exec("DELETE FROM " + table); err != nil {
			return fmt.Errorf("failed to clear %s: %w", table, err)
		}
	}

	// Insert new data
	if err := insertNoteFolders(tx, data.NoteFolders); err != nil {
		return err
	}
	if err := insertNotes(tx, data.Notes); err != nil {
		return err
	}
	if err := insertTodoFolders(tx, data.TodoFolders); err != nil {
		return err
	}
	if err := insertTodos(tx, data.Todos); err != nil {
		return err
	}
	if err := insertTodoSteps(tx, data.TodoSteps); err != nil {
		return err
	}

	return tx.Commit()
}

// Insert helpers (using transaction)
func insertNoteFolders(tx *sql.Tx, folders []NoteFolderSync) error {
	stmt, err := tx.Prepare(`INSERT INTO note_folders (id, name, parent_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, f := range folders {
		_, err := stmt.Exec(f.ID, f.Name, f.ParentID, f.CreatedAt, f.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertNotes(tx *sql.Tx, notes []NoteSync) error {
	stmt, err := tx.Prepare(`INSERT INTO notes (id, title, content, folder_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, n := range notes {
		_, err := stmt.Exec(n.ID, n.Title, n.Content, n.FolderID, n.CreatedAt, n.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertTodoFolders(tx *sql.Tx, folders []TodoFolderSync) error {
	stmt, err := tx.Prepare(`INSERT INTO todo_folders (id, name, parent_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, f := range folders {
		_, err := stmt.Exec(f.ID, f.Name, f.ParentID, f.CreatedAt, f.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertTodos(tx *sql.Tx, todos []TodoSync) error {
	stmt, err := tx.Prepare(`INSERT INTO todos (id, title, folder_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, t := range todos {
		_, err := stmt.Exec(t.ID, t.Title, t.FolderID, t.CreatedAt, t.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertTodoSteps(tx *sql.Tx, steps []TodoStepSync) error {
	stmt, err := tx.Prepare(`INSERT INTO todo_steps (id, todo_id, description, completed, display_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, s := range steps {
		_, err := stmt.Exec(s.ID, s.TodoID, s.Description, s.Completed, s.DisplayOrder, s.CreatedAt, s.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}
