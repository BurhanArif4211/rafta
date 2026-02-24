package repository

import (
	"database/sql"
	"github.com/burhanarif4211/rafta/internal/models"
	_ "modernc.org/sqlite"
	"testing"
)

// setupTestDB creates an in-memory SQLite database with the full schema.
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		t.Fatal(err)
	}

	// Execute schema (same as in db package)
	schema := `
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

CREATE INDEX IF NOT EXISTS idx_note_folders_parent ON note_folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_notes_folder ON notes(folder_id);
CREATE INDEX IF NOT EXISTS idx_todo_folders_parent ON todo_folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_todos_folder ON todos(folder_id);
CREATE INDEX IF NOT EXISTS idx_todo_steps_todo ON todo_steps(todo_id);
`
	if _, err := db.Exec(schema); err != nil {
		t.Fatal(err)
	}
	return db
}

// Helper to create a null string
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// ---------- NoteFolder tests ----------
func TestNoteFolderRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewNoteFolderRepository(db)

	// Create root folder
	root := models.NewNoteFolder("Root", nullString(""))
	if err := repo.Create(root); err != nil {
		t.Fatal(err)
	}

	// Create child folder
	child := models.NewNoteFolder("Child", nullString(root.ID))
	if err := repo.Create(child); err != nil {
		t.Fatal(err)
	}

	// Get roots
	roots, err := repo.GetRoots()
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 || roots[0].ID != root.ID {
		t.Errorf("expected 1 root, got %d", len(roots))
	}

	// Get children
	children, err := repo.GetChildren(root.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 1 || children[0].ID != child.ID {
		t.Errorf("expected 1 child, got %d", len(children))
	}

	// Update folder
	child.Name = "Renamed Child"
	if err := repo.Update(child); err != nil {
		t.Fatal(err)
	}
	updated, err := repo.GetByID(child.ID)
	if err != nil || updated.Name != "Renamed Child" {
		t.Errorf("update failed, got name %s", updated.Name)
	}

	// Delete child
	if err := repo.Delete(child.ID); err != nil {
		t.Fatal(err)
	}
	_, err = repo.GetByID(child.ID)
	if err != sql.ErrNoRows {
		t.Error("child should be deleted")
	}
}

// ---------- Note tests ----------
func TestNoteRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	folderRepo := NewNoteFolderRepository(db)
	noteRepo := NewNoteRepository(db)

	// Create a folder
	folder := models.NewNoteFolder("Notes", nullString(""))
	if err := folderRepo.Create(folder); err != nil {
		t.Fatal(err)
	}

	// Create two notes
	note1 := models.NewNote("Note 1", "Content 1", folder.ID)
	note2 := models.NewNote("Note 2", "Content 2", folder.ID)
	if err := noteRepo.Create(note1); err != nil {
		t.Fatal(err)
	}
	if err := noteRepo.Create(note2); err != nil {
		t.Fatal(err)
	}

	// Get by folder
	notes, err := noteRepo.GetByFolder(folder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}

	// Update note
	note1.Title = "Updated Note 1"
	if err := noteRepo.Update(note1); err != nil {
		t.Fatal(err)
	}
	updated, err := noteRepo.GetByID(note1.ID)
	if err != nil || updated.Title != "Updated Note 1" {
		t.Error("update failed")
	}

	// Delete note
	if err := noteRepo.Delete(note1.ID); err != nil {
		t.Fatal(err)
	}
	_, err = noteRepo.GetByID(note1.ID)
	if err != sql.ErrNoRows {
		t.Error("note should be deleted")
	}

	// Delete folder should cascade delete note2
	if err := folderRepo.Delete(folder.ID); err != nil {
		t.Fatal(err)
	}
	notes, err = noteRepo.GetByFolder(folder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Error("cascade delete failed")
	}
}

// ---------- TodoFolder tests ----------
func TestTodoFolderRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewTodoFolderRepository(db)

	root := models.NewTodoFolder("Todo Root", nullString(""))
	if err := repo.Create(root); err != nil {
		t.Fatal(err)
	}
	child := models.NewTodoFolder("Todo Child", nullString(root.ID))
	if err := repo.Create(child); err != nil {
		t.Fatal(err)
	}

	roots, err := repo.GetRoots()
	if err != nil || len(roots) != 1 {
		t.Error("root count mismatch")
	}
	children, err := repo.GetChildren(root.ID)
	if err != nil || len(children) != 1 {
		t.Error("children count mismatch")
	}

	// Delete
	if err := repo.Delete(child.ID); err != nil {
		t.Fatal(err)
	}
	_, err = repo.GetByID(child.ID)
	if err != sql.ErrNoRows {
		t.Error("child should be deleted")
	}
}

// ---------- Todo tests ----------
func TestTodoRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	folderRepo := NewTodoFolderRepository(db)
	todoRepo := NewTodoRepository(db)

	folder := models.NewTodoFolder("Todo Folder", nullString(""))
	if err := folderRepo.Create(folder); err != nil {
		t.Fatal(err)
	}

	todo1 := models.NewTodo("Task 1", folder.ID)
	todo2 := models.NewTodo("Task 2", folder.ID)
	if err := todoRepo.Create(todo1); err != nil {
		t.Fatal(err)
	}
	if err := todoRepo.Create(todo2); err != nil {
		t.Fatal(err)
	}

	todos, err := todoRepo.GetByFolder(folder.ID)
	if err != nil || len(todos) != 2 {
		t.Error("expected 2 todos")
	}

	// Update
	todo1.Title = "Updated Task"
	if err := todoRepo.Update(todo1); err != nil {
		t.Fatal(err)
	}
	updated, _ := todoRepo.GetByID(todo1.ID)
	if updated.Title != "Updated Task" {
		t.Error("update failed")
	}

	// Delete folder cascade
	if err := folderRepo.Delete(folder.ID); err != nil {
		t.Fatal(err)
	}
	todos, _ = todoRepo.GetByFolder(folder.ID)
	if len(todos) != 0 {
		t.Error("cascade delete failed")
	}
}

// ---------- TodoStep tests ----------
func TestTodoStepRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	folderRepo := NewTodoFolderRepository(db)
	todoRepo := NewTodoRepository(db)
	stepRepo := NewTodoStepRepository(db)

	folder := models.NewTodoFolder("Folder", nullString(""))
	if err := folderRepo.Create(folder); err != nil {
		t.Fatal(err)
	}
	todo := models.NewTodo("Main Todo", folder.ID)
	if err := todoRepo.Create(todo); err != nil {
		t.Fatal(err)
	}

	step1 := models.NewTodoStep(todo.ID, "Step 1", 1)
	step2 := models.NewTodoStep(todo.ID, "Step 2", 2)
	step3 := models.NewTodoStep(todo.ID, "Step 3", 3)
	for _, s := range []*models.TodoStep{step1, step2, step3} {
		if err := stepRepo.Create(s); err != nil {
			t.Fatal(err)
		}
	}

	// Get by todo
	steps, err := stepRepo.GetByTodo(todo.ID)
	if err != nil || len(steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(steps))
	}

	// Update step
	step1.Description = "Updated Step 1"
	step1.Completed = true
	if err := stepRepo.Update(step1); err != nil {
		t.Fatal(err)
	}
	updated, _ := stepRepo.GetByID(step1.ID)
	if updated.Description != "Updated Step 1" || !updated.Completed {
		t.Error("update failed")
	}

	// Reorder steps: put step3 first, step2 second, step1 third
	err = stepRepo.Reorder(todo.ID, []string{step3.ID, step2.ID, step1.ID})
	if err != nil {
		t.Fatal(err)
	}
	steps, _ = stepRepo.GetByTodo(todo.ID)
	if steps[0].ID != step3.ID || steps[1].ID != step2.ID || steps[2].ID != step1.ID {
		t.Error("reorder failed")
	}
	// Verify display_order values
	if steps[0].DisplayOrder != 1 || steps[1].DisplayOrder != 2 || steps[2].DisplayOrder != 3 {
		t.Error("display_order not updated correctly")
	}

	// Delete step
	if err := stepRepo.Delete(step1.ID); err != nil {
		t.Fatal(err)
	}
	steps, _ = stepRepo.GetByTodo(todo.ID)
	if len(steps) != 2 {
		t.Error("step delete failed")
	}

	// Delete todo should cascade
	if err := todoRepo.Delete(todo.ID); err != nil {
		t.Fatal(err)
	}
	steps, _ = stepRepo.GetByTodo(todo.ID)
	if len(steps) != 0 {
		t.Error("cascade delete from todo failed")
	}
}

// ---------- Device tests ----------
func TestDeviceRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewDeviceRepository(db)

	dev := models.NewDevice("192.168.0.10", "laptop")
	if err := repo.Create(dev); err != nil {
		t.Fatal(err)
	}

	// Get by ID
	fetched, err := repo.GetByID(dev.ID)
	if err != nil || fetched.LocalIP != dev.LocalIP {
		t.Error("get by ID failed")
	}

	// Get by IP
	fetchedByIP, err := repo.GetByIP("192.168.0.10")
	if err != nil || fetchedByIP.ID != dev.ID {
		t.Error("get by IP failed")
	}

	// Update
	dev.Hostname = "new-laptop"
	if err := repo.Update(dev); err != nil {
		t.Fatal(err)
	}
	updated, _ := repo.GetByID(dev.ID)
	if updated.Hostname != "new-laptop" {
		t.Error("update failed")
	}

	// UpdateLastSeen
	if err := repo.UpdateLastSeen(dev.ID); err != nil {
		t.Fatal(err)
	}
	// just check no error; we could verify timestamp changed but not critical

	// GetAll
	all, err := repo.GetAll()
	if err != nil || len(all) != 1 {
		t.Error("get all failed")
	}

	// Delete
	if err := repo.Delete(dev.ID); err != nil {
		t.Fatal(err)
	}
	_, err = repo.GetByID(dev.ID)
	if err != sql.ErrNoRows {
		t.Error("delete failed")
	}
}
