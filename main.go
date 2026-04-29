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
	cors := middleware.CORS(cfg.CORS.AllowedOrigins)
	log := middleware.Logging(cfg.Debug)
	timeout := middleware.Timeout(10)

	common := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.Chain(h,
			middleware.Recovery,
			middleware.RequestID,
			cors,
			log,
			timeout,
		)
	}

	// Users
	mux.HandleFunc("GET /users", common(userHandler.GetAll))
	mux.HandleFunc("POST /users", common(userHandler.Create))

	// Projects
	mux.HandleFunc("GET /projects", common(projectHandler.GetAll))
	mux.HandleFunc("POST /projects", common(projectHandler.Create))

	// Tasks
	mux.HandleFunc("GET /users/{userId}/tasks", common(taskHandler.GetByUser))
	mux.HandleFunc("GET /projects/{projectId}/tasks", common(taskHandler.GetByProject))
	mux.HandleFunc("POST /tasks", common(taskHandler.Create))
	mux.HandleFunc("PATCH /tasks/{id}", common(taskHandler.UpdateStatus))
	mux.HandleFunc("DELETE /tasks/{id}", common(taskHandler.Delete))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server started on %s\n", addr)
	_ = http.ListenAndServe(addr, mux)
	return
}
