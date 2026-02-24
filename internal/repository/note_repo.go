package repository

import (
	"database/sql"
	"github.com/burhanarif4211/rafta/internal/models"
	"time"
)

type NoteRepository interface {
	Create(note *models.Note) error
	GetAll() ([]*models.Note, error)
	GetByID(id string) (*models.Note, error)
	GetByFolder(folderID string) ([]*models.Note, error)
	Update(note *models.Note) error
	Delete(id string) error
}

type noteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) NoteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) Create(note *models.Note) error {
	query := `INSERT INTO notes (id, title, content, folder_id, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, note.ID, note.Title, note.Content, note.FolderID, note.CreatedAt, note.UpdatedAt)
	return err
}

func (r *noteRepository) GetAll() ([]*models.Note, error) {
	rows, err := r.db.Query(`SELECT id, title, content, folder_id, created_at, updated_at FROM notes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []*models.Note
	for rows.Next() {
		var n models.Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.FolderID, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, &n)
	}
	return notes, rows.Err()
}

func (r *noteRepository) GetByID(id string) (*models.Note, error) {
	var n models.Note
	query := `SELECT id, title, content, folder_id, created_at, updated_at FROM notes WHERE id = ?`
	row := r.db.QueryRow(query, id)
	err := row.Scan(&n.ID, &n.Title, &n.Content, &n.FolderID, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *noteRepository) GetByFolder(folderID string) ([]*models.Note, error) {
	rows, err := r.db.Query(`SELECT id, title, content, folder_id, created_at, updated_at FROM notes WHERE folder_id = ?`, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []*models.Note
	for rows.Next() {
		var n models.Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.FolderID, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, &n)
	}
	return notes, rows.Err()
}

func (r *noteRepository) Update(note *models.Note) error {
	note.UpdatedAt = time.Now()
	query := `UPDATE notes SET title = ?, content = ?, folder_id = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, note.Title, note.Content, note.FolderID, note.UpdatedAt, note.ID)
	return err
}

func (r *noteRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM notes WHERE id = ?`, id)
	return err
}
