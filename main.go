package main

import (
	"fmt"
	"net/http"
	"os"
	"taskmanager/config"
	"taskmanager/handler"
	"taskmanager/middleware"
	"taskmanager/store"

	_ "modernc.org/sqlite"
)

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := store.InitDB(cfg.Database.Path)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userStore := store.NewUserStore(db)
	userHandler := handler.NewUserHandler(userStore)
	projectHandler := handler.NewProjectHandler(store.NewProjectStore(db), userStore)
	taskHandler := handler.NewTaskHandler(store.NewTaskStore(db))

	mux := http.NewServeMux()

	// Users
	mux.HandleFunc("GET /users", middleware.Logging(userHandler.GetAll))
	mux.HandleFunc("POST /users", middleware.Logging(userHandler.Create))

	// Projects
	mux.HandleFunc("GET /projects", middleware.Logging(projectHandler.GetAll))
	mux.HandleFunc("POST /projects", middleware.Logging(projectHandler.Create))

	// Tasks
	mux.HandleFunc("GET /users/{userId}/tasks", middleware.Logging(taskHandler.GetByUser))
	mux.HandleFunc("GET /projects/{projectId}/tasks", middleware.Logging(taskHandler.GetByProject))
	mux.HandleFunc("POST /tasks", middleware.Logging(taskHandler.Create))
	mux.HandleFunc("PATCH /tasks/{id}", middleware.Logging(taskHandler.UpdateStatus))
	mux.HandleFunc("DELETE /tasks/{id}", middleware.Logging(taskHandler.Delete))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server started on %s\n", addr)
	_ = http.ListenAndServe(addr, mux)
	return
}
