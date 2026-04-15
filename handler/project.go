package handler

import (
	"encoding/json"
	"net/http"
	"taskmanager/model"
	"taskmanager/store"
)

type ProjectHandler struct {
	Store     *store.ProjectStore
	UserStore *store.UserStore
}

func NewProjectHandler(s *store.ProjectStore, us *store.UserStore) *ProjectHandler {
	return &ProjectHandler{Store: s, UserStore: us}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Name == "" {
		errorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.OwnerID == 0 {
		errorResponse(w, http.StatusBadRequest, "Owner id is required")
		return
	}
	_, err := h.UserStore.GetByID(req.OwnerID)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Owner not found")
		return
	}

	task, err := h.Store.Create(req.Name, req.Description, req.OwnerID)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, task)
}

func (h *ProjectHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	projects, err := h.Store.GetAll()
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, projects)
}
