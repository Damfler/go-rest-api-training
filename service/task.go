package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	logger      *slog.Logger
}

func NewTaskService(repo TaskRepository, projectRepo ProjectRepository, logger *slog.Logger) *TaskService {
	return &TaskService{
		repo:        repo,
		projectRepo: projectRepo,
		logger:      logger.With("service", "task"),
	}
}

var validStatuses = map[string]bool{
	"todo": true, "in_progress": true, "done": true,
}

func (s *TaskService) Create(ctx context.Context, req model.CreateTaskRequest) (*model.Task, error) {
	s.logger.Info("creating task", "projectId", req.ProjectID)

	if err := s.validateCreate(req); err != nil {
		s.logger.Warn("validation failed", "err", err)
		return nil, err
	}

	_, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		s.logger.Error("failed to create task", "err", err)
		return nil, fmt.Errorf("project not found: %w", err)
	}

	task, err := s.repo.Create(ctx, req.Title, req.ProjectID, req.UserID)
	if err != nil {
		s.logger.Error("failed to create task", "err", err)
		return nil, err
	}

	s.logger.Info("task created", "taskID", task.ID)
	return task, nil
}

func (s *TaskService) UpdateStatus(ctx context.Context, id int, status string) error {
	s.logger.Info("updating status", "id", id)

	if !validStatuses[status] {
		return &apperror.ValidationError{Field: "status", Message: "must be todo, in_progress or done"}
	}

	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *TaskService) GetByProject(ctx context.Context, projectID int, status string) ([]model.Task, error) {
	s.logger.Info("getting tasks", "projectId", projectID, "status", status)

	if status != "" && !validStatuses[status] {
		return nil, &apperror.ValidationError{Field: "status", Message: "must be todo, in_progress or done"}
	}

	return s.repo.GetByProject(ctx, projectID, status)
}

func (s *TaskService) GetByUser(ctx context.Context, userID int, status string) ([]model.Task, error) {
	s.logger.Info("getting tasks", "userId", userID, "status", status)
	return s.repo.GetByUser(ctx, userID, status)
}

func (s *TaskService) Delete(ctx context.Context, id int) error {
	s.logger.Info("deleting task", "id", id)
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
