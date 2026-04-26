package service

import (
	"context"
	"errors"
	"strings"
	"taskmanager/apperror"
	"taskmanager/model"
)

type UserRepository interface {
	Create(ctx context.Context, name, email string) (*model.User, error)
	GetAll(ctx context.Context) ([]model.User, error)
	GetByID(ctx context.Context, id int) (*model.User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	if err := s.validateCreate(req); err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	return s.repo.Create(ctx, name, email)
}

func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetAll(ctx context.Context) ([]model.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) validateCreate(req model.CreateUserRequest) error {
	var errs []error

	if strings.TrimSpace(req.Name) == "" {
		errs = append(errs, &apperror.ValidationError{Field: "name", Message: "required"})
	}

	if !strings.Contains(req.Email, "@") {
		errs = append(errs, &apperror.ValidationError{Field: "email", Message: "must contain @"})
	}

	return errors.Join(errs...)
}
