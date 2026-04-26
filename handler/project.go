package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"taskmanager/apperror"
	"taskmanager/model"
)

type ProjectService interface {
	Create(ctx context.Context, req model.CreateProjectRequest) (*model.Project, error)
	GetAll(ctx context.Context) ([]model.Project, error)
	GetByID(ctx context.Context, id int) (*model.Project, error)
}

type ProjectHandler struct {
	service ProjectService
}

func NewProjectHandler(s ProjectService) *ProjectHandler {
	return &ProjectHandler{service: s}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	task, err := h.service.Create(ctx, req)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, task)
}

func (h *ProjectHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projects, err := h.service.GetAll(ctx)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, projects)
}
