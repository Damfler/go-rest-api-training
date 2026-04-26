package model

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"` // "todo", "in_progress", "done"
	ProjectID int    `json:"project_id"`
	UserID    *int   `json:"user_id,omitempty"` // может быть не назначен
}

type CreateTaskRequest struct {
	Title     string `json:"title"`
	ProjectID int    `json:"project_id"`
	UserID    *int   `json:"user_id,omitempty"`
}

type UpdateTaskRequest struct {
	Status string `json:"status"`
	UserID *int   `json:"user_id,omitempty"`
}

func (t Task) GetID() int {
	return t.ID
}
