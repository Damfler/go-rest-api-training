package store

import (
	"database/sql"
	"fmt"
	"taskmanager/model"
)

type TaskStore struct {
	DB *sql.DB
}

var validStatus = map[string]bool{"todo": true, "in_progress": true, "done": true}

func NewTaskStore(db *sql.DB) *TaskStore {
	return &TaskStore{DB: db}
}

func (s *TaskStore) Create(title string, projectID int, userID *int) (*model.Task, error) {
	result, err := s.DB.Exec(
		"INSERT INTO tasks (title, project_id, user_id) VALUES (?, ?, ?)",
		title, projectID, userID,
	)
	if err != nil {
		return nil, err
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

func (s *TaskStore) getTasksWhere(column string, value int, status string) ([]model.Task, error) {
	query := "SELECT id, title, status, project_id, user_id FROM tasks WHERE " + column + " = ?"
	args := []any{value}

	if status != "" {
		if !validStatus[status] {
			return nil, fmt.Errorf("invalid status: %s", status)
		}
		query += " AND status = ?"
		args = append(args, status)
	}

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.ProjectID, &t.UserID)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (s *TaskStore) GetByProject(projectID int, status string) ([]model.Task, error) {
	return s.getTasksWhere("project_id", projectID, status)
}

func (s *TaskStore) GetByUser(userID int, status string) ([]model.Task, error) {
	return s.getTasksWhere("user_id", userID, status)
}

func (s *TaskStore) UpdateStatus(id int, status string) error {
	if !validStatus[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	result, err := s.DB.Exec("UPDATE tasks SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task %d not found", id)
	}

	return nil
}

func (s *TaskStore) Assign(taskID int, userID *int) error {
	result, err := s.DB.Exec("UPDATE tasks SET user_id = ? WHERE id = ?", userID, taskID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task %d not found", taskID)
	}

	return nil
}

func (s *TaskStore) Delete(id int) error {
	result, err := s.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task %d not found", id)
	}

	return nil
}
