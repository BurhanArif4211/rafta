package repository

import (
	"database/sql"
	"github.com/burhanarif4211/rafta/internal/models"
	"time"
)

type TodoStepRepository interface {
	Create(step *models.TodoStep) error
	GetByID(id string) (*models.TodoStep, error)
	GetByTodo(todoID string) ([]*models.TodoStep, error)
	Update(step *models.TodoStep) error
	Delete(id string) error
	// convenience: reorder steps (bulk update of display_order)
	Reorder(todoID string, stepIDs []string) error
}

type todoStepRepository struct {
	db *sql.DB
}

func NewTodoStepRepository(db *sql.DB) TodoStepRepository {
	return &todoStepRepository{db: db}
}

func (r *todoStepRepository) Create(step *models.TodoStep) error {
	query := `INSERT INTO todo_steps (id, todo_id, description, completed, display_order, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, step.ID, step.TodoID, step.Description, step.Completed, step.DisplayOrder, step.CreatedAt, step.UpdatedAt)
	return err
}

func (r *todoStepRepository) GetByID(id string) (*models.TodoStep, error) {
	var s models.TodoStep
	query := `SELECT id, todo_id, description, completed, display_order, created_at, updated_at FROM todo_steps WHERE id = ?`
	row := r.db.QueryRow(query, id)
	err := row.Scan(&s.ID, &s.TodoID, &s.Description, &s.Completed, &s.DisplayOrder, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *todoStepRepository) GetByTodo(todoID string) ([]*models.TodoStep, error) {
	rows, err := r.db.Query(`SELECT id, todo_id, description, completed, display_order, created_at, updated_at FROM todo_steps WHERE todo_id = ? ORDER BY display_order`, todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var steps []*models.TodoStep
	for rows.Next() {
		var s models.TodoStep
		if err := rows.Scan(&s.ID, &s.TodoID, &s.Description, &s.Completed, &s.DisplayOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		steps = append(steps, &s)
	}
	return steps, rows.Err()
}

func (r *todoStepRepository) Update(step *models.TodoStep) error {
	step.UpdatedAt = time.Now()
	query := `UPDATE todo_steps SET description = ?, completed = ?, display_order = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, step.Description, step.Completed, step.DisplayOrder, step.UpdatedAt, step.ID)
	return err
}

func (r *todoStepRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM todo_steps WHERE id = ?`, id)
	return err
}

func (r *todoStepRepository) Reorder(todoID string, stepIDs []string) error {
	// stepIDs are in the new order; we assign display_order = index+1 for each step.
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, stepID := range stepIDs {
		newOrder := i + 1
		_, err := tx.Exec(`UPDATE todo_steps SET display_order = ?, updated_at = ? WHERE id = ? AND todo_id = ?`, newOrder, time.Now(), stepID, todoID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
