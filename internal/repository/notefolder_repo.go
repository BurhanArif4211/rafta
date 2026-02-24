package repository

import (
	"database/sql"
	"github.com/burhanarif4211/rafta/internal/models"
	"time"
)

type NoteFolderRepository interface {
	Create(folder *models.NoteFolder) error
	GetAll() ([]*models.NoteFolder, error)
	GetByID(id string) (*models.NoteFolder, error)
	GetRoots() ([]*models.NoteFolder, error)
	GetChildren(parentID string) ([]*models.NoteFolder, error)
	Update(folder *models.NoteFolder) error
	Delete(id string) error
}

type noteFolderRepository struct {
	db *sql.DB
}

func NewNoteFolderRepository(db *sql.DB) NoteFolderRepository {
	return &noteFolderRepository{db: db}
}

func (r *noteFolderRepository) Create(folder *models.NoteFolder) error {
	query := `INSERT INTO note_folders (id, name, parent_id, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, folder.ID, folder.Name, folder.ParentID, folder.CreatedAt, folder.UpdatedAt)
	return err
}

func (r *noteFolderRepository) GetAll() ([]*models.NoteFolder, error) {
	rows, err := r.db.Query(`SELECT id, name, parent_id, created_at, updated_at FROM note_folders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var folders []*models.NoteFolder
	for rows.Next() {
		var f models.NoteFolder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, &f)
	}
	return folders, rows.Err()
}

func (r *noteFolderRepository) GetByID(id string) (*models.NoteFolder, error) {
	var f models.NoteFolder
	query := `SELECT id, name, parent_id, created_at, updated_at FROM note_folders WHERE id = ?`
	row := r.db.QueryRow(query, id)
	err := row.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *noteFolderRepository) GetRoots() ([]*models.NoteFolder, error) {
	rows, err := r.db.Query(`SELECT id, name, parent_id, created_at, updated_at FROM note_folders WHERE parent_id IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var folders []*models.NoteFolder
	for rows.Next() {
		var f models.NoteFolder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, &f)
	}
	return folders, rows.Err()
}

func (r *noteFolderRepository) GetChildren(parentID string) ([]*models.NoteFolder, error) {
	rows, err := r.db.Query(`SELECT id, name, parent_id, created_at, updated_at FROM note_folders WHERE parent_id = ?`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var folders []*models.NoteFolder
	for rows.Next() {
		var f models.NoteFolder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, &f)
	}
	return folders, rows.Err()
}

func (r *noteFolderRepository) Update(folder *models.NoteFolder) error {
	folder.UpdatedAt = time.Now()
	query := `UPDATE note_folders SET name = ?, parent_id = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, folder.Name, folder.ParentID, folder.UpdatedAt, folder.ID)
	return err
}

func (r *noteFolderRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM note_folders WHERE id = ?`, id)
	return err
}
