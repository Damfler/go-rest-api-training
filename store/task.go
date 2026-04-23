package store

import (
	"context"
	"database/sql"
	"fmt"
	"taskmanager/apperror"
	"taskmanager/model"
)

type TaskStore struct {
	DB *sql.DB
}

var validStatus = map[string]bool{"todo": true, "in_progress": true, "done": true}

func NewTaskStore(db *sql.DB) *TaskStore {
	return &TaskStore{DB: db}
}

func (s *TaskStore) Create(ctx context.Context, title string, projectID int, userID *int) (*model.Task, error) {
	result, err := s.DB.ExecContext(ctx,
		"INSERT INTO tasks (title, project_id, user_id) VALUES (?, ?, ?)",
		title, projectID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("TaskStore.Create(%s,%d,%d): %w", title, projectID, userID, err)
	}

	id, _ := result.LastInsertId()
	return &model.Task{
		ID:        int(id),
		Title:     title,
		Status:    "todo",
		ProjectID: projectID,
		UserID:    userID,
	}, nil
}

func (s *TaskStore) getTasksWhere(ctx context.Context, column string, value int, status string) ([]model.Task, error) {
	query := "SELECT id, title, status, project_id, user_id FROM tasks WHERE " + column + " = ?"
	args := []any{value}

	if status != "" {
		if !validStatus[status] {
			return nil, &apperror.ValidationError{Field: "status", Message: "must be todo, in_progress or done"}
		}
		query += " AND status = ?"
		args = append(args, status)
	}

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("TaskStore.getTasksWhere(%s,%d,%s): %w", column, value, status, err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.ProjectID, &t.UserID)
		if err != nil {
			return nil, fmt.Errorf("TaskStore.getTasksWhere scan: %w", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (s *TaskStore) GetByProject(ctx context.Context, projectID int, status string) ([]model.Task, error) {
	return s.getTasksWhere(ctx, "project_id", projectID, status)
}

func (s *TaskStore) GetByUser(ctx context.Context, userID int, status string) ([]model.Task, error) {
	return s.getTasksWhere(ctx, "user_id", userID, status)
}

func (s *TaskStore) UpdateStatus(ctx context.Context, id int, status string) error {
	if !validStatus[status] {
		return &apperror.ValidationError{Field: "status", Message: "must be todo, in_progress or done"}
	}

	result, err := s.DB.ExecContext(ctx, "UPDATE tasks SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("TaskStore.UpdateStatus(%s,%d): %w", status, id, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &apperror.NotFoundError{Entity: "task", ID: id}
	}

	return nil
}

func (s *TaskStore) Assign(ctx context.Context, taskID int, userID *int) error {
	result, err := s.DB.ExecContext(ctx, "UPDATE tasks SET user_id = ? WHERE id = ?", userID, taskID)
	if err != nil {
		return fmt.Errorf("TaskStore.Assign(%d,%d): %w", taskID, userID, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &apperror.NotFoundError{Entity: "task", ID: taskID}
	}

	return nil
}

func (s *TaskStore) Delete(ctx context.Context, id int) error {
	result, err := s.DB.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("TaskStore.Delete(%d): %w", id, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &apperror.NotFoundError{Entity: "task", ID: id}
	}

	return nil
}
