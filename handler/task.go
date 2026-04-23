package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"taskmanager/apperror"
	"taskmanager/model"
)

type TaskStore interface {
	Create(ctx context.Context, title string, projectID int, userID *int) (*model.Task, error)
	GetByProject(ctx context.Context, projectID int, status string) ([]model.Task, error)
	GetByUser(ctx context.Context, userID int, status string) ([]model.Task, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	Delete(ctx context.Context, id int) error
}

type TaskHandler struct {
	Store TaskStore
}

func NewTaskHandler(s TaskStore) *TaskHandler {
	return &TaskHandler{Store: s}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := validateCreateTask(req); err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	task, err := h.Store.Create(ctx, req.Title, req.ProjectID, req.UserID)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetByProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID, err := strconv.Atoi(r.PathValue("projectId"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	status := r.URL.Query().Get("status")

	tasks, err := h.Store.GetByProject(ctx, projectID, status)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := strconv.Atoi(r.PathValue("userId"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	status := r.URL.Query().Get("status")

	tasks, err := h.Store.GetByUser(ctx, userID, status)
	if err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req model.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := h.Store.UpdateStatus(ctx, id, req.Status); err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.Store.Delete(ctx, id); err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func validateCreateTask(req model.CreateTaskRequest) error {
	var errs []error

	if req.Title == "" {
		errs = append(errs, &apperror.ValidationError{
			Field: "title", Message: "required",
		})
	}
	if req.ProjectID == 0 {
		errs = append(errs, &apperror.ValidationError{
			Field: "project_id", Message: "required",
		})
	}

	return errors.Join(errs...)
}
