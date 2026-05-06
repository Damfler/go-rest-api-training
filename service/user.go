package service

import (
	"context"
	"errors"
	"log/slog"
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
	repo   UserRepository
	logger *slog.Logger
}

func NewUserService(repo UserRepository, logger *slog.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger.With("service", "user"),
	}
}

func (s *UserService) Create(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	s.logger.Info("creating user", "name", req.Name, "email", req.Email)

	if err := s.validateCreate(req); err != nil {
		s.logger.Warn("validation failed", "err", err)
		return nil, err
	}

	user, err := s.repo.Create(ctx, req.Name, req.Email)
	if err != nil {
		s.logger.Error("failed to create user", "err", err)
		return nil, err
	}

	s.logger.Info("user created", "userID", user.ID)
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, error) {
	s.logger.Info("getting user", "id", id)
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetAll(ctx context.Context) ([]model.User, error) {
	s.logger.Info("get all user")
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
