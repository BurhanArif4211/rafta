package repository

import (
	"database/sql"
	"time"

	"github.com/burhanarif4211/rafta/internal/models"
)

type TodoFolderRepository interface {
	Create(folder *models.TodoFolder) error
	GetAll() ([]*models.TodoFolder, error)
	GetByID(id string) (*models.TodoFolder, error)
	GetRoots() ([]*models.TodoFolder, error)
	GetChildren(parentID string) ([]*models.TodoFolder, error)
	Update(folder *models.TodoFolder) error
	Delete(id string) error
}

type todoFolderRepository struct {
	db *sql.DB
}

func NewTodoFolderRepository(db *sql.DB) TodoFolderRepository {
	return &todoFolderRepository{db: db}
}

func (r *todoFolderRepository) Create(folder *models.TodoFolder) error {
	query := `INSERT INTO todo_folders (id, name, parent_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, folder.ID, folder.Name, folder.ParentID, folder.CreatedAt, folder.UpdatedAt)
	return err
}

func (r *todoFolderRepository) GetAll() ([]*models.TodoFolder, error) {
	rows, err := r.db.Query(`SELECT id, name, parent_id, created_at, updated_at FROM todo_folders ORDER BY updated_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var todofolders []*models.TodoFolder
	for rows.Next() {
		var n models.TodoFolder
		if err := rows.Scan(&n.ID, &n.Name, &n.ParentID, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		todofolders = append(todofolders, &n)
	}
	return todofolders, rows.Err()
}

func (r *todoFolderRepository) GetByID(id string) (*models.TodoFolder, error) {
	var f models.TodoFolder
	query := `SELECT id, name, parent_id, created_at, updated_at FROM todo_folders WHERE id = ?`
	row := r.db.QueryRow(query, id)
	err := row.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *todoFolderRepository) GetRoots() ([]*models.TodoFolder, error) {
	rows, err := r.db.Query(`SELECT id, name, parent_id, created_at, updated_at FROM todo_folders WHERE parent_id IS NULL ORDER BY updated_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var folders []*models.TodoFolder
	for rows.Next() {
		var f models.TodoFolder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, &f)
	}
	return folders, rows.Err()
}

func (r *todoFolderRepository) GetChildren(parentID string) ([]*models.TodoFolder, error) {
	rows, err := r.db.Query(`SELECT id, name, parent_id, created_at, updated_at FROM todo_folders WHERE parent_id = ?`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var folders []*models.TodoFolder
	for rows.Next() {
		var f models.TodoFolder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, &f)
	}
	return folders, rows.Err()
}

func (r *todoFolderRepository) Update(folder *models.TodoFolder) error {
	folder.UpdatedAt = time.Now()
	query := `UPDATE todo_folders SET name = ?, parent_id = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, folder.Name, folder.ParentID, folder.UpdatedAt, folder.ID)
	return err
}

func (r *todoFolderRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM todo_folders WHERE id = ?`, id)
	return err
}
