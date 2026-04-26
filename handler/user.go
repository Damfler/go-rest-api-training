package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"taskmanager/apperror"
	"taskmanager/model"
)

type UserService interface {
	Create(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetByID(ctx context.Context, id int) (*model.User, error)
	GetAll(ctx context.Context) ([]model.User, error)
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(s UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	user, err := h.service.Create(r.Context(), req)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, user)
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAll(r.Context())
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, users)
}
