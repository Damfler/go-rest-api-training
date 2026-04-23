package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"taskmanager/apperror"
	"taskmanager/model"
	"taskmanager/store"
	"testing"

	_ "modernc.org/sqlite"
)

type testEnv struct {
	UserStore      *store.UserStore
	ProjectStore   *store.ProjectStore
	TaskStore      *store.TaskStore
	UserHandler    *UserHandler
	ProjectHandler *ProjectHandler
	TaskHandler    *TaskHandler
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE
        );
        CREATE TABLE projects (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            description TEXT,
            owner_id INTEGER NOT NULL,
            FOREIGN KEY (owner_id) REFERENCES users(id)
        );
        CREATE TABLE tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            status TEXT DEFAULT 'todo',
            project_id INTEGER NOT NULL,
            user_id INTEGER,
            FOREIGN KEY (project_id) REFERENCES projects(id),
            FOREIGN KEY (user_id) REFERENCES users(id)
        );
    `)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	db := setupTestDB(t)

	us := store.NewUserStore(db)
	ps := store.NewProjectStore(db)
	ts := store.NewTaskStore(db)

	return &testEnv{
		UserStore:      us,
		ProjectStore:   ps,
		TaskStore:      ts,
		UserHandler:    NewUserHandler(us),
		ProjectHandler: NewProjectHandler(ps, us),
		TaskHandler:    NewTaskHandler(ts),
	}
}

func TestCreateUser(t *testing.T) {
	env := setupTestEnv(t)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedName   string
	}{
		{
			name:           "success",
			body:           `{"name":"Alex","email":"alex@mail.com"}`,
			expectedStatus: http.StatusCreated,
			expectedName:   "Alex",
		},
		{
			name:           "empty name",
			body:           `{"name":"","email":"alex@mail.com"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty email",
			body:           `{"name":"Alex","email":""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			body:           `{broken`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/users",
				bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			env.UserHandler.Create(w, req)

			// Проверяем статус
			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			// Для успешных — проверяем тело
			if tt.expectedStatus == http.StatusCreated {
				var user model.User
				json.NewDecoder(w.Body).Decode(&user)

				if user.Name != tt.expectedName {
					t.Errorf("name = %q, want %q", user.Name, tt.expectedName)
				}
				if user.ID == 0 {
					t.Error("expected user ID to be set")
				}
			}
		})
	}
}

func TestGetAllUsers(t *testing.T) {
	env := setupTestEnv(t)

	users := []string{
		`{"name":"Alex","email":"alex@mail.com"}`,
		`{"name":"Bob","email":"bob@mail.com"}`,
	}
	for _, body := range users {
		req := httptest.NewRequest("POST", "/users",
			bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.UserHandler.Create(w, req)
	}

	// Запрашиваем список
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	env.UserHandler.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var result []model.User
	json.NewDecoder(w.Body).Decode(&result)

	if len(result) != 2 {
		t.Errorf("got %d users, want 2", len(result))
	}
}

func TestCreateProject(t *testing.T) {
	env := setupTestEnv(t)

	user, err := env.UserStore.Create(t.Context(), "Alex", "alex@mail.com")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedName   string
	}{
		{
			name:           "success",
			body:           `{"name":"Rest","description":"","owner_id":` + strconv.Itoa(user.ID) + `}`,
			expectedStatus: http.StatusCreated,
			expectedName:   "Rest",
		},
		{
			name:           "empty name",
			body:           `{"name":"","description":"","owner_id":` + strconv.Itoa(user.ID) + `}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty owner id",
			body:           `{"name":"Alex","description":"","owner_id":0}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			body:           `{broken`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/projects",
				bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			env.ProjectHandler.Create(w, req)

			// Проверяем статус
			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			// Для успешных — проверяем тело
			if tt.expectedStatus == http.StatusCreated {
				var project model.Project
				json.NewDecoder(w.Body).Decode(&project)

				if project.Name != tt.expectedName {
					t.Errorf("name = %q, want %q", project.Name, tt.expectedName)
				}
				if project.ID == 0 {
					t.Error("expected user ID to be set")
				}
			}
		})
	}
}

func TestGetAllProjects(t *testing.T) {
	env := setupTestEnv(t)

	user, _ := env.UserStore.Create(t.Context(), "Alex", "alex@mail.com")

	projects := []string{
		`{"name":"Rest","description":"","owner_id":` + strconv.Itoa(user.ID) + `}`,
		`{"name":"Home2","description":"","owner_id":` + strconv.Itoa(user.ID) + `}`,
	}
	for _, body := range projects {
		req := httptest.NewRequest("POST", "/projects",
			bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.ProjectHandler.Create(w, req)
	}

	// Запрашиваем список
	req := httptest.NewRequest("GET", "/projects", nil)
	w := httptest.NewRecorder()
	env.ProjectHandler.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var result []model.Project
	json.NewDecoder(w.Body).Decode(&result)

	if len(result) != 2 {
		t.Errorf("got %d projects, want 2", len(result))
	}
}

func TestTaskProject(t *testing.T) {
	env := setupTestEnv(t)

	// Подготовка — реальный пользователь и проект
	user, _ := env.UserStore.Create(t.Context(), "Alex", "alex@mail.com")
	project, _ := env.ProjectStore.Create(t.Context(), "Project 1", "desc", user.ID)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedTitle  string
	}{
		{
			name:           "success",
			body:           `{"title":"Testing","project_id":` + strconv.Itoa(project.ID) + `,"user_id":` + strconv.Itoa(user.ID) + `}`,
			expectedStatus: http.StatusCreated,
			expectedTitle:  "Testing",
		},
		{
			name:           "empty title",
			body:           `{"title":"","project_id":` + strconv.Itoa(project.ID) + `}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty project id",
			body:           `{"title":"Testing","project_id":0}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			body:           `{broken`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/tasks",
				bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			env.TaskHandler.Create(w, req)

			// Проверяем статус
			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			// Для успешных — проверяем тело
			if tt.expectedStatus == http.StatusCreated {
				var task model.Task
				json.NewDecoder(w.Body).Decode(&task)
				if task.Title != tt.expectedTitle {
					t.Errorf("title = %q, want %q", task.Title, tt.expectedTitle)
				}
			}
		})
	}
}

func TestTaskGetByProject(t *testing.T) {
	db := setupTestDB(t)

	userStore := store.NewUserStore(db)
	projectStore := store.NewProjectStore(db)
	taskStore := store.NewTaskStore(db)
	taskHandler := NewTaskHandler(taskStore)

	user, err := userStore.Create(t.Context(), "test", "test@mail.com")
	if err != nil {
		t.Fatal(err)
	}

	var project1 model.Project
	for i := 1; i <= 2; i++ {
		project, err := projectStore.Create(t.Context(), "Project "+strconv.Itoa(i), "desc", user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if i == 1 {
			project1 = *project
		}

		taskStore.Create(t.Context(), "Task "+strconv.Itoa(i), project.ID, nil)
	}

	req := httptest.NewRequest("GET",
		"/projects/"+strconv.Itoa(project1.ID)+"/tasks", nil)
	req.SetPathValue("projectId", strconv.Itoa(project1.ID))
	w := httptest.NewRecorder()

	taskHandler.GetByProject(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var tasks []model.Task
	json.NewDecoder(w.Body).Decode(&tasks)

	if len(tasks) != 1 {
		t.Errorf("got %d tasks, want 1", len(tasks))
	}

	for _, task := range tasks {
		if task.ProjectID != project1.ID {
			t.Errorf("task project_id = %d, want %d",
				task.ProjectID, project1.ID)
		}
	}
}

func TestTaskGetByProjectAndByUser(t *testing.T) {
	db := setupTestDB(t)

	userStore := store.NewUserStore(db)
	projectStore := store.NewProjectStore(db)
	taskStore := store.NewTaskStore(db)
	taskHandler := NewTaskHandler(taskStore)

	user, err := userStore.Create(t.Context(), "test", "test@mail.com")
	if err != nil {
		t.Fatal(err)
	}

	var project1 model.Project
	for i := 1; i <= 2; i++ {
		project, err := projectStore.Create(t.Context(), "Project "+strconv.Itoa(i), "desc", user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if i == 1 {
			project1 = *project
		}

		taskStore.Create(t.Context(), "Task "+strconv.Itoa(i), project.ID, &user.ID)
	}

	req := httptest.NewRequest("GET",
		"/projects/"+strconv.Itoa(project1.ID)+"/tasks", nil)
	req.SetPathValue("projectId", strconv.Itoa(project1.ID))
	w := httptest.NewRecorder()

	taskHandler.GetByProject(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var tasks []model.Task
	json.NewDecoder(w.Body).Decode(&tasks)

	if len(tasks) != 1 {
		t.Errorf("got %d tasks, want 1", len(tasks))
	}

	for _, task := range tasks {
		if task.ProjectID != project1.ID {
			t.Errorf("task project_id = %d, want %d",
				task.ProjectID, project1.ID)
		}
	}

	req = httptest.NewRequest("GET",
		"/users/"+strconv.Itoa(user.ID)+"/tasks", nil)
	req.SetPathValue("userId", strconv.Itoa(user.ID))
	w = httptest.NewRecorder()

	taskHandler.GetByUser(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	tasks = []model.Task{}
	json.NewDecoder(w.Body).Decode(&tasks)

	if len(tasks) != 2 {
		t.Errorf("got %d tasks, want 2", len(tasks))
	}
}

func TestUpdateStatus(t *testing.T) {
	env := setupTestEnv(t)

	user, _ := env.UserStore.Create(t.Context(), "Alex", "alex@mail.com")
	project, _ := env.ProjectStore.Create(t.Context(), "Project", "desc", user.ID)
	task, _ := env.TaskStore.Create(t.Context(), "Fix bug", project.ID, nil)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{"valid status", `{"status":"in_progress"}`, http.StatusOK},
		{"done", `{"status":"done"}`, http.StatusOK},
		{"invalid status", `{"status":"cancelled"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("PATCH",
				"/tasks/"+strconv.Itoa(task.ID),
				bytes.NewReader([]byte(tt.body)))
			req.SetPathValue("id", strconv.Itoa(task.ID))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			env.TaskHandler.UpdateStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	env := setupTestEnv(t)

	user, _ := env.UserStore.Create(t.Context(), "Alex", "alex@mail.com")
	project, _ := env.ProjectStore.Create(t.Context(), "Project", "desc", user.ID)
	task, _ := env.TaskStore.Create(t.Context(), "Delete me", project.ID, nil)

	req := httptest.NewRequest("DELETE",
		"/tasks/"+strconv.Itoa(task.ID), nil)
	req.SetPathValue("id", strconv.Itoa(task.ID))
	w := httptest.NewRecorder()

	env.TaskHandler.Delete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("delete status = %d, want 200", w.Code)
	}

	req = httptest.NewRequest("DELETE",
		"/tasks/"+strconv.Itoa(task.ID), nil)
	req.SetPathValue("id", strconv.Itoa(task.ID))
	w = httptest.NewRecorder()

	env.TaskHandler.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("second delete status = %d, want 404", w.Code)
	}
}

/* Mock */
type mockTaskStore struct {
	tasks []model.Task
	err   error
}

func (m *mockTaskStore) Create(ctx context.Context, title string, projectID int, userID *int) (*model.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	task := &model.Task{
		ID:        len(m.tasks) + 1,
		Title:     title,
		ProjectID: projectID,
		UserID:    userID,
		Status:    "todo",
	}
	m.tasks = append(m.tasks, *task)
	return task, nil
}

func (m *mockTaskStore) GetByProject(ctx context.Context, projectID int, status string) ([]model.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []model.Task
	for _, t := range m.tasks {
		if t.ProjectID == projectID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskStore) GetByUser(ctx context.Context, userID int, status string) ([]model.Task, error) {
	return nil, nil
}

func (m *mockTaskStore) UpdateStatus(ctx context.Context, id int, status string) error {
	return m.err
}

func (m *mockTaskStore) Delete(ctx context.Context, id int) error {
	return m.err
}

func TestCreateTaskHandler(t *testing.T) {
	mock := &mockTaskStore{}
	h := &TaskHandler{Store: mock}

	body := bytes.NewReader([]byte(`{"title":"Test","project_id":1}`))
	req := httptest.NewRequest("POST", "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}

	// Проверяем что мок получил данные
	if len(mock.tasks) != 1 {
		t.Errorf("got %d tasks in store, want 1", len(mock.tasks))
	}
}

func TestCreateTaskDBError(t *testing.T) {
	mock := &mockTaskStore{err: fmt.Errorf("database is down")}
	h := &TaskHandler{Store: mock}

	body := bytes.NewReader([]byte(`{"title":"Test","project_id":1}`))
	req := httptest.NewRequest("POST", "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestRemoveTaskDBError(t *testing.T) {
	mock := &mockTaskStore{err: &apperror.NotFoundError{Entity: "task", ID: 999}}
	h := &TaskHandler{Store: mock}

	req := httptest.NewRequest("DELETE", "/tasks/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

type mockUserStore struct {
	users []model.User
	err   error
}

func (m *mockUserStore) Create(ctx context.Context, name, email string) (*model.User, error) {
	if m.err != nil {
		return nil, m.err
	}

	for _, u := range m.users {
		if u.Email == email {
			return nil, &apperror.ConflictError{
				Entity: "user", Field: "email", Value: email,
			}
		}
	}

	user := &model.User{
		ID:    len(m.users) + 1,
		Name:  name,
		Email: email,
	}
	m.users = append(m.users, *user)
	return user, nil
}

func (s *mockUserStore) GetAll(ctx context.Context) ([]model.User, error) {
	return nil, nil
}

func (s *mockUserStore) GetByID(ctx context.Context, id int) (*model.User, error) {
	return nil, nil
}

func TestCreateDoubleUsersHandler(t *testing.T) {
	mock := &mockUserStore{}
	h := &UserHandler{Store: mock}

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "success",
			body:           `{"name":"Test","email":"test@example.com"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "double email",
			body:           `{"name":"Test","email":"test@example.com"}`,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewReader([]byte(tt.body))
			req := httptest.NewRequest("POST", "/users", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Create(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}
