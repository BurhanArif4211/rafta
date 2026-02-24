package repository

import (
	"database/sql"
	"github.com/burhanarif4211/rafta/internal/models"
	"time"
)

type TodoRepository interface {
	Create(todo *models.Todo) error
	GetAll() ([]*models.Todo, error)
	GetByID(id string) (*models.Todo, error)
	GetByFolder(folderID string) ([]*models.Todo, error)
	Update(todo *models.Todo) error
	Delete(id string) error
}

type todoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(todo *models.Todo) error {
	query := `INSERT INTO todos (id, title, folder_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, todo.ID, todo.Title, todo.FolderID, todo.CreatedAt, todo.UpdatedAt)
	return err
}

func (r *todoRepository) GetAll() ([]*models.Todo, error) {
	rows, err := r.db.Query(`SELECT id, title, folder_id, created_at, updated_at FROM todos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var todos []*models.Todo
	for rows.Next() {
		var n models.Todo
		if err := rows.Scan(&n.ID, &n.Title, &n.FolderID, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, &n)
	}
	return todos, rows.Err()
}

func (r *todoRepository) GetByID(id string) (*models.Todo, error) {
	var t models.Todo
	query := `SELECT id, title, folder_id, created_at, updated_at FROM todos WHERE id = ?`
	row := r.db.QueryRow(query, id)
	err := row.Scan(&t.ID, &t.Title, &t.FolderID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *todoRepository) GetByFolder(folderID string) ([]*models.Todo, error) {
	rows, err := r.db.Query(`SELECT id, title, folder_id, created_at, updated_at FROM todos WHERE folder_id = ?`, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var todos []*models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.FolderID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, &t)
	}
	return todos, rows.Err()
}

func (r *todoRepository) Update(todo *models.Todo) error {
	todo.UpdatedAt = time.Now()
	query := `UPDATE todos SET title = ?, folder_id = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, todo.Title, todo.FolderID, todo.UpdatedAt, todo.ID)
	return err
}

func (r *todoRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM todos WHERE id = ?`, id)
	return err
}
