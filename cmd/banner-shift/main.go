package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/JustForWorld/banner-shift/internal/config"
	delete_banner "github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/delete"
	"github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/get"
	getlist "github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/get-list"
	"github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/save"
	"github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/update"
	"github.com/JustForWorld/banner-shift/internal/storage/postgresql"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting banner-shift", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := postgresql.New(
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.DB,
		cfg.Port,
	)
	if err != nil {
		slog.Error("failed to init storage: %w", err)
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/banner", getlist.New(log, storage))
	router.Get("/user_banner", get.New(log, storage))

	router.Post("/banner", save.New(log, storage))
	router.Patch("/banner/{id}", update.New(log, storage))
	router.Delete("/banner/{id}", delete_banner.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))
	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}
	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		))
	}

	return log
}
