package sync

import "time"

type NoteFolderSync struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parent_id"` // nil means NULL
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NoteSync struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	FolderID  string    `json:"folder_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoFolderSync struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoSync struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	FolderID  string    `json:"folder_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoStepSync struct {
	ID           string    `json:"id"`
	TodoID       string    `json:"todo_id"`
	Description  string    `json:"description"`
	Completed    bool      `json:"completed"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Combined data for a full sync
type SyncData struct {
	NoteFolders []NoteFolderSync `json:"note_folders"`
	Notes       []NoteSync       `json:"notes"`
	TodoFolders []TodoFolderSync `json:"todo_folders"`
	Todos       []TodoSync       `json:"todos"`
	TodoSteps   []TodoStepSync   `json:"todo_steps"`
}
