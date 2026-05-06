package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"taskmanager/model"
	"taskmanager/repository"
	"taskmanager/service"
	"testing"

	_ "modernc.org/sqlite"
)

type testApp struct {
	UserHandler    *UserHandler
	ProjectHandler *ProjectHandler
	TaskHandler    *TaskHandler
	DB             *sql.DB
}

func setupIntegration(t *testing.T) *testApp {
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	userService := service.NewUserService(userRepo, logger)
	projectService := service.NewProjectService(projectRepo, userRepo, logger)
	taskService := service.NewTaskService(taskRepo, projectRepo, logger)

	t.Cleanup(func() { db.Close() })

	return &testApp{
		UserHandler:    NewUserHandler(userService),
		ProjectHandler: NewProjectHandler(projectService),
		TaskHandler:    NewTaskHandler(taskService),
		DB:             db,
	}
}

func TestFullWorkflow(t *testing.T) {
	app := setupIntegration(t)

	// 1. Создаём пользователя
	body := bytes.NewReader([]byte(`{"name":"Alex","email":"alex@mail.com"}`))
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.UserHandler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create user: status = %d, want 201", w.Code)
	}

	var user model.User
	json.NewDecoder(w.Body).Decode(&user)
	t.Logf("Created user: %+v", user)

	// 2. Создаём проект
	projectBody := `{"name":"My Project","description":"test","owner_id":` + strconv.Itoa(user.ID) + `}`
	req = httptest.NewRequest("POST", "/projects", bytes.NewReader([]byte(projectBody)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	app.ProjectHandler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create project: status = %d, want 201", w.Code)
	}

	var project model.Project
	json.NewDecoder(w.Body).Decode(&project)
	t.Logf("Created project: %+v", project)

	// 3. Создаём задачу
	taskBody := `{"title":"Fix bug","project_id":` + strconv.Itoa(project.ID) + `}`
	req = httptest.NewRequest("POST", "/tasks", bytes.NewReader([]byte(taskBody)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	app.TaskHandler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create task: status = %d, want 201", w.Code)
	}

	var task model.Task
	json.NewDecoder(w.Body).Decode(&task)
	t.Logf("Created task: %+v", task)

	if task.Status != "todo" {
		t.Errorf("new task status = %q, want todo", task.Status)
	}

	// 4. Меняем статус
	req = httptest.NewRequest("PATCH", "/tasks/"+strconv.Itoa(task.ID),
		bytes.NewReader([]byte(`{"status":"done"}`)))
	req.SetPathValue("id", strconv.Itoa(task.ID))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	app.TaskHandler.UpdateStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update status: status = %d, want 200", w.Code)
	}

	// 5. Проверяем задачи проекта
	req = httptest.NewRequest("GET", "/projects/"+strconv.Itoa(project.ID)+"/tasks", nil)
	req.SetPathValue("projectId", strconv.Itoa(project.ID))
	w = httptest.NewRecorder()
	app.TaskHandler.GetByProject(w, req)

	var tasks []model.Task
	json.NewDecoder(w.Body).Decode(&tasks)

	if len(tasks) != 1 {
		t.Errorf("got %d tasks, want 1", len(tasks))
	}

	// 6. Удаляем задачу
	req = httptest.NewRequest("DELETE", "/tasks/"+strconv.Itoa(task.ID), nil)
	req.SetPathValue("id", strconv.Itoa(task.ID))
	w = httptest.NewRecorder()
	app.TaskHandler.Delete(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("delete: status = %d, want 200", w.Code)
	}

	// 7. Проверяем что задач больше нет
	req = httptest.NewRequest("GET", "/projects/"+strconv.Itoa(project.ID)+"/tasks", nil)
	req.SetPathValue("projectId", strconv.Itoa(project.ID))
	w = httptest.NewRecorder()
	app.TaskHandler.GetByProject(w, req)

	tasks = nil
	json.NewDecoder(w.Body).Decode(&tasks)

	if len(tasks) != 0 {
		t.Errorf("got %d tasks after delete, want 0", len(tasks))
	}

	t.Log("Full workflow passed!")
}

func TestValidationErrors(t *testing.T) {
	app := setupIntegration(t)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{"empty user name", "POST", "/users", `{"name":"","email":"test@mail.com"}`, 400},
		{"invalid email", "POST", "/users", `{"name":"Test","email":"invalid"}`, 400},
		{"empty project name", "POST", "/projects", `{"name":"","owner_id":1}`, 400},
		{"empty task title", "POST", "/tasks", `{"title":"","project_id":1}`, 400},
		{"invalid json", "POST", "/users", `{broken`, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path,
				bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			switch tt.path {
			case "/users":
				app.UserHandler.Create(w, req)
			case "/projects":
				app.ProjectHandler.Create(w, req)
			case "/tasks":
				app.TaskHandler.Create(w, req)
			}

			if w.Code != tt.status {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.status, w.Body.String())
			}
		})
	}
}
