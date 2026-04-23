package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"taskmanager/apperror"
	"taskmanager/model"
)

type UserGetter interface {
	GetByID(ctx context.Context, id int) (*model.User, error)
}

type ProjectStore interface {
	Create(ctx context.Context, name, description string, ownerID int) (*model.Project, error)
	GetAll(ctx context.Context) ([]model.Project, error)
	GetByID(ctx context.Context, id int) (*model.Project, error)
}

type ProjectHandler struct {
	Store     ProjectStore
	UserStore UserGetter
}

func NewProjectHandler(s ProjectStore, us UserGetter) *ProjectHandler {
	return &ProjectHandler{Store: s, UserStore: us}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := validateCreateProject(req); err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	_, err := h.UserStore.GetByID(ctx, req.OwnerID)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), "Owner not found")
		return
	}

	task, err := h.Store.Create(ctx, req.Name, req.Description, req.OwnerID)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, task)
}

func (h *ProjectHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projects, err := h.Store.GetAll(ctx)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, projects)
}

func validateCreateProject(req model.CreateProjectRequest) error {
	var errs []error

	if req.Name == "" {
		errs = append(errs, &apperror.ValidationError{
			Field: "name", Message: "required",
		})
	}
	if req.OwnerID == 0 {
		errs = append(errs, &apperror.ValidationError{
			Field: "owner_id", Message: "required",
		})
	}

	return errors.Join(errs...)
}
