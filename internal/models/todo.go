package models

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type TodoFolder struct {
	ID        string
	Name      string
	ParentID  sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTodoFolder(name string, parentID sql.NullString) *TodoFolder {
	now := time.Now()
	return &TodoFolder{
		ID:        uuid.New().String(),
		Name:      name,
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type Todo struct {
	ID        string
	Title     string
	FolderID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTodo(title, folderID string) *Todo {
	now := time.Now()
	return &Todo{
		ID:        uuid.New().String(),
		Title:     title,
		FolderID:  folderID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type TodoStep struct {
	ID           string
	TodoID       string
	Description  string
	Completed    bool
	DisplayOrder int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewTodoStep(todoID, description string, order int) *TodoStep {
	now := time.Now()
	return &TodoStep{
		ID:           uuid.New().String(),
		TodoID:       todoID,
		Description:  description,
		Completed:    false,
		DisplayOrder: order,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
