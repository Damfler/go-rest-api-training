package service

import (
	"context"
	"errors"
	"fmt"
	"taskmanager/apperror"
	"taskmanager/model"
)

type TaskRepository interface {
	Create(ctx context.Context, title string, projectID int, userID *int) (*model.Task, error)
	GetByProject(ctx context.Context, projectID int, status string) ([]model.Task, error)
	GetByUser(ctx context.Context, userID int, status string) ([]model.Task, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	Delete(ctx context.Context, id int) error
}

type TaskService struct {
	repo        TaskRepository
	projectRepo ProjectRepository
}

func NewTaskService(repo TaskRepository, projectRepo ProjectRepository) *TaskService {
	return &TaskService{repo: repo, projectRepo: projectRepo}
}

var validStatuses = map[string]bool{
	"todo": true, "in_progress": true, "done": true,
}

func (s *TaskService) Create(ctx context.Context, req model.CreateTaskRequest) (*model.Task, error) {
	if err := s.validateCreate(req); err != nil {
		return nil, err
	}

	// Бизнес-правило: проект должен существовать
	_, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.repo.Create(ctx, req.Title, req.ProjectID, req.UserID)
}

func (s *TaskService) UpdateStatus(ctx context.Context, id int, status string) error {
	if !validStatuses[status] {
		return &apperror.ValidationError{Field: "status", Message: "must be todo, in_progress or done"}
	}

	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *TaskService) GetByProject(ctx context.Context, projectID int, status string) ([]model.Task, error) {
	if status != "" && !validStatuses[status] {
		return nil, &apperror.ValidationError{Field: "status", Message: "must be todo, in_progress or done"}
	}

	return s.repo.GetByProject(ctx, projectID, status)
}

func (s *TaskService) GetByUser(ctx context.Context, userID int, status string) ([]model.Task, error) {
	return s.repo.GetByUser(ctx, userID, status)
}

func (s *TaskService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *TaskService) validateCreate(req model.CreateTaskRequest) error {
	var errs []error

	if req.Title == "" {
		errs = append(errs, &apperror.ValidationError{Field: "title", Message: "required"})
	}
	if req.ProjectID == 0 {
		errs = append(errs, &apperror.ValidationError{Field: "project_id", Message: "required"})
	}

	return errors.Join(errs...)
}
