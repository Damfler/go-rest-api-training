package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"taskmanager/apperror"
	"taskmanager/model"
)

type UserStore interface {
	Create(ctx context.Context, name, email string) (*model.User, error)
	GetAll(ctx context.Context) ([]model.User, error)
	GetByID(ctx context.Context, id int) (*model.User, error)
}

type UserHandler struct {
	Store UserStore
}

func NewUserHandler(s UserStore) *UserHandler {
	return &UserHandler{Store: s}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := validateCreateUser(req); err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	task, err := h.Store.Create(ctx, req.Name, req.Email)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, task)
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.Store.GetAll(ctx)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, users)
}

func validateCreateUser(req model.CreateUserRequest) error {
	var errs []error

	if req.Name == "" {
		errs = append(errs, &apperror.ValidationError{
			Field: "name", Message: "required",
		})
	}
	if !strings.Contains(req.Email, "@") {
		errs = append(errs, &apperror.ValidationError{
			Field: "email", Message: "must contain @",
		})
	}

	return errors.Join(errs...)
}
