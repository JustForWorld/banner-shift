package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/JustForWorld/banner-shift/internal/config"
	"github.com/JustForWorld/banner-shift/internal/storage/postgresql"
	"github.com/go-chi/chi/v5"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Test OK!</h1>")
}

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
	_ = storage

	_ = chi.NewRouter()

	http.HandleFunc("/", testHandler)
	http.ListenAndServe(":8080", nil)
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
