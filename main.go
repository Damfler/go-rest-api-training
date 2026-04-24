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
