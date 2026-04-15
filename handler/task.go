package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"taskmanager/model"
	"taskmanager/store"
)

type TaskHandler struct {
	Store *store.TaskStore
}

func NewTaskHandler(s *store.TaskStore) *TaskHandler {
	return &TaskHandler{Store: s}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Title == "" {
		errorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}

	if req.ProjectID == 0 {
		errorResponse(w, http.StatusBadRequest, "Project id is required")
		return
	}

	task, err := h.Store.Create(req.Title, req.ProjectID, req.UserID)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetByProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(r.PathValue("projectId"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	status := r.URL.Query().Get("status")

	tasks, err := h.Store.GetByProject(projectID, status)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("userId"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	status := r.URL.Query().Get("status")

	tasks, err := h.Store.GetByUser(userID, status)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	if err := h.Store.UpdateStatus(id, req.Status); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.Store.Delete(id); err != nil {
		errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "deleted"})
}
