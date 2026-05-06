package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"taskmanager/apperror"
	"taskmanager/model"
)

type ProjectRepository interface {
	Create(ctx context.Context, name, description string, ownerId int) (*model.Project, error)
	GetAll(ctx context.Context) ([]model.Project, error)
	GetByID(ctx context.Context, id int) (*model.Project, error)
}

type ProjectService struct {
	repo     ProjectRepository
	userRepo UserRepository
	logger   *slog.Logger
}

func NewProjectService(repo ProjectRepository, userRepo UserRepository, logger *slog.Logger) *ProjectService {
	return &ProjectService{
		repo:     repo,
		userRepo: userRepo,
		logger:   logger.With("service", "project"),
	}
}

func (s *ProjectService) Create(ctx context.Context, req model.CreateProjectRequest) (*model.Project, error) {
	s.logger.Info("create project", "name", req.Name, "description", req.Description, "ownerId", req.OwnerID)

	if err := s.validateCreate(req); err != nil {
		s.logger.Warn("validation failed", "error", err)
		return nil, err
	}

	_, err := s.userRepo.GetByID(ctx, req.OwnerID)
	if err != nil {
		s.logger.Error("user does not exist", "error", err)
		return nil, fmt.Errorf("owner not found: %w", err)
	}

	project, err := s.repo.Create(ctx, req.Name, req.Description, req.OwnerID)
	if err != nil {
		s.logger.Error("failed to create project", "err", err)
		return nil, err
	}

	s.logger.Info("project created", "projectID", project.ID)
	return project, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id int) (*model.Project, error) {
	s.logger.Info("get project", "id", id)
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) GetAll(ctx context.Context) ([]model.Project, error) {
	s.logger.Info("get all projects")
	return s.repo.GetAll(ctx)
}

func (s *ProjectService) validateCreate(req model.CreateProjectRequest) error {
	var errs []error

	if strings.TrimSpace(req.Name) == "" {
		errs = append(errs, &apperror.ValidationError{Field: "name", Message: "required"})
	}

	return errors.Join(errs...)
}
