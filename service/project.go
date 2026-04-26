package service

import (
	"context"
	"errors"
	"fmt"
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
}

func NewProjectService(repo ProjectRepository, userRepo UserRepository) *ProjectService {
	return &ProjectService{repo: repo, userRepo: userRepo}
}

func (s *ProjectService) Create(ctx context.Context, req model.CreateProjectRequest) (*model.Project, error) {
	if err := s.validateCreate(req); err != nil {
		return nil, err
	}

	_, err := s.userRepo.GetByID(ctx, req.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("owner not found: %w", err)
	}

	return s.repo.Create(ctx, req.Name, req.Description, req.OwnerID)
}

func (s *ProjectService) GetByID(ctx context.Context, id int) (*model.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) GetAll(ctx context.Context) ([]model.Project, error) {
	return s.repo.GetAll(ctx)
}

func (s *ProjectService) validateCreate(req model.CreateProjectRequest) error {
	var errs []error

	if strings.TrimSpace(req.Name) == "" {
		errs = append(errs, &apperror.ValidationError{Field: "name", Message: "required"})
	}

	return errors.Join(errs...)
}
