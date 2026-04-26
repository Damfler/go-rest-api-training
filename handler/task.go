package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"taskmanager/apperror"
	"taskmanager/model"
)

type TaskService interface {
	Create(ctx context.Context, req model.CreateTaskRequest) (*model.Task, error)
	GetByProject(ctx context.Context, projectID int, status string) ([]model.Task, error)
	GetByUser(ctx context.Context, userID int, status string) ([]model.Task, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	Delete(ctx context.Context, id int) error
}

type TaskHandler struct {
	service TaskService
}

func NewTaskHandler(s TaskService) *TaskHandler {
	return &TaskHandler{service: s}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.CreateTaskRequest
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

func (h *TaskHandler) GetByProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID, err := strconv.Atoi(r.PathValue("projectId"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	status := r.URL.Query().Get("status")

	tasks, err := h.service.GetByProject(ctx, projectID, status)
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

	tasks, err := h.service.GetByUser(ctx, userID, status)
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

	if err := h.service.UpdateStatus(ctx, id, req.Status); err != nil {
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

	if err := h.service.Delete(ctx, id); err != nil {
		errorResponse(w, apperror.HTTPStatus(err), err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "deleted"})
}
