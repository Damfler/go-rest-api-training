package main

import (
	"fmt"
	"net/http"
	"taskmanager/handler"
	"taskmanager/middleware"
	"taskmanager/store"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := store.InitDB("app.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userHandler := handler.NewUserHandler(store.NewUserStore(db))
	projectHandler := handler.NewProjectHandler(store.NewProjectStore(db), store.NewUserStore(db))
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

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", mux)
}
