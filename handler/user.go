package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"taskmanager/model"
)

type UserStore interface {
	Create(name, email string) (*model.User, error)
	GetAll() ([]model.User, error)
	GetByID(id int) (*model.User, error)
}

type UserHandler struct {
	Store UserStore
}

func NewUserHandler(s UserStore) *UserHandler {
	return &UserHandler{Store: s}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Name == "" {
		errorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}

	if !strings.Contains(req.Email, "@") {
		errorResponse(w, http.StatusBadRequest, "Email is required")
		return
	}

	task, err := h.Store.Create(req.Name, req.Email)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, task)
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.Store.GetAll()
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, users)
}
