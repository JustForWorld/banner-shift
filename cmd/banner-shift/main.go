package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/JustForWorld/banner-shift/internal/config"
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

	// TODO: storage: postresql

	// TODO: router: go-chi

	// TODO: run server
	// fmt.Println("test")

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
