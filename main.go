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

	// Users
	mux.HandleFunc("GET /users", log(userHandler.GetAll))
	mux.HandleFunc("POST /users", log(userHandler.Create))

	// Projects
	mux.HandleFunc("GET /projects", log(projectHandler.GetAll))
	mux.HandleFunc("POST /projects", log(projectHandler.Create))

	// Tasks
	mux.HandleFunc("GET /users/{userId}/tasks", log(taskHandler.GetByUser))
	mux.HandleFunc("GET /projects/{projectId}/tasks", log(taskHandler.GetByProject))
	mux.HandleFunc("POST /tasks", log(taskHandler.Create))
	mux.HandleFunc("PATCH /tasks/{id}", log(taskHandler.UpdateStatus))
	mux.HandleFunc("DELETE /tasks/{id}", log(taskHandler.Delete))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server started on %s\n", addr)
	_ = http.ListenAndServe(addr, mux)
	return
}
