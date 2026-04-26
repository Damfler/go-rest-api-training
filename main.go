package main

import (
	"fmt"
	"net/http"
	"taskmanager/config"
	"taskmanager/handler"
	"taskmanager/middleware"
	"taskmanager/repository"
	"taskmanager/service"

	_ "modernc.org/sqlite"
)

func main() {
	cfg, _ := config.Load("config.yaml")
	db, _ := repository.InitDB(cfg.Database.Path)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	userService := service.NewUserService(userRepo)
	projectService := service.NewProjectService(projectRepo, userRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo)

	userHandler := handler.NewUserHandler(userService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)

	mux := http.NewServeMux()
	log := middleware.Logging(cfg.Debug)
	timeout := middleware.Timeout(10)

	// Users
	mux.HandleFunc("GET /users", log(timeout(userHandler.GetAll)))
	mux.HandleFunc("POST /users", log(timeout(userHandler.Create)))

	// Projects
	mux.HandleFunc("GET /projects", log(timeout(projectHandler.GetAll)))
	mux.HandleFunc("POST /projects", log(timeout(projectHandler.Create)))

	// Tasks
	mux.HandleFunc("GET /users/{userId}/tasks", log(timeout(taskHandler.GetByUser)))
	mux.HandleFunc("GET /projects/{projectId}/tasks", log(timeout(taskHandler.GetByProject)))
	mux.HandleFunc("POST /tasks", log(timeout(taskHandler.Create)))
	mux.HandleFunc("PATCH /tasks/{id}", log(timeout(taskHandler.UpdateStatus)))
	mux.HandleFunc("DELETE /tasks/{id}", log(timeout(taskHandler.Delete)))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server started on %s\n", addr)
	_ = http.ListenAndServe(addr, mux)
	return
}
