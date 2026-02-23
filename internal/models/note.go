package models

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type NoteFolder struct {
	ID        string
	Name      string
	ParentID  sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewNoteFolder(name string, parentID sql.NullString) *NoteFolder {
	now := time.Now()
	return &NoteFolder{
		ID:        uuid.New().String(),
		Name:      name,
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type Note struct {
	ID        string
	Title     string
	Content   string
	FolderID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewNote(title, content, folderID string) *Note {
	now := time.Now()
	return &Note{
		ID:        uuid.New().String(),
		Title:     title,
		Content:   content,
		FolderID:  folderID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
