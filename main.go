package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskmanager/config"
	"taskmanager/handler"
	loggerPack "taskmanager/logger"
	"taskmanager/middleware"
	"taskmanager/repository"
	"taskmanager/service"

	_ "modernc.org/sqlite"
)

func main() {
	// Config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}

	// Logger
	logger := loggerPack.Setup(cfg.Debug)
	logger.Info("config loaded", "debug", cfg.Debug)

	// Database
	db, err := repository.InitDB(cfg.Database.Path)
	if err != nil {
		logger.Error("database connection failed", "err", err)
		os.Exit(1)
	}
	logger.Info("database connected", "path", cfg.Database.Path)

	// Repository
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Service
	userService := service.NewUserService(userRepo, logger)
	projectService := service.NewProjectService(projectRepo, userRepo, logger)
	taskService := service.NewTaskService(taskRepo, projectRepo, logger)

	// Handler
	userHandler := handler.NewUserHandler(userService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)

	// Router
	mux := http.NewServeMux()
	cors := middleware.CORS(cfg.CORS.AllowedOrigins)
	log := middleware.Logging(logger)
	timeout := middleware.Timeout(10)

	common := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.Chain(h,
			middleware.Recovery(logger),
			middleware.RequestID,
			cors,
			log,
			timeout,
		)
	}

	mux.HandleFunc("GET /users", common(userHandler.GetAll))
	mux.HandleFunc("POST /users", common(userHandler.Create))
	mux.HandleFunc("GET /projects", common(projectHandler.GetAll))
	mux.HandleFunc("POST /projects", common(projectHandler.Create))
	mux.HandleFunc("GET /users/{userId}/tasks", common(taskHandler.GetByUser))
	mux.HandleFunc("GET /projects/{projectId}/tasks", common(taskHandler.GetByProject))
	mux.HandleFunc("POST /tasks", common(taskHandler.Create))
	mux.HandleFunc("PATCH /tasks/{id}", common(taskHandler.UpdateStatus))
	mux.HandleFunc("DELETE /tasks/{id}", common(taskHandler.Delete))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		logger.Info("server started", "addr", addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// Ждём сигнал остановки
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("shutdown signal received", "signal", sig.String())

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced shutdown", "err", err)
	}

	// Закрываем ресурсы
	db.Close()

	logger.Info("shutdown complete")
}
