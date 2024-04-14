package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/JustForWorld/banner-shift/internal/config"
	delete_banner "github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/delete"
	"github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/get"
	getlist "github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/get-list"
	"github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/save"
	"github.com/JustForWorld/banner-shift/internal/http-server/handlers/banner/update"
	"github.com/JustForWorld/banner-shift/internal/storage/postgresql"
	"github.com/JustForWorld/banner-shift/internal/storage/redis"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var (
	tokenAuth *jwtauth.JWTAuth
)

func init() {
	tokenAuth = jwtauth.New("HS256", []byte(fmt.Sprint(time.Now().Unix())), nil)
}

func main() {
	cfg, usr := config.MustLoad()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	log := setupLogger(cfg.Env)
	log.Info("starting banner-shift", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"username": usr.Username, "role": usr.Role, "tag": usr.Tag})
	log.Debug("current jwt token", slog.String("jwt", tokenString))

	redis, err := redis.New(
		ctx,
		cfg.Redis.Addr,
		cfg.Redis.User,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.Protocol,
	)
	if err != nil {
		slog.Error("failed to init Redis storage: %w", err)
		os.Exit(1)
	}

	storage, err := postgresql.New(
		ctx,
		cfg.PostgreSQL.User,
		cfg.PostgreSQL.Password,
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.DB,
		cfg.PostgreSQL.Port,
	)
	if err != nil {
		slog.Error("failed to init PostgreSQL storage: %w", err)
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Use(jwtauth.Verifier(tokenAuth))
	router.Use(jwtauth.Authenticator(tokenAuth))

	router.Get("/banner", getlist.New(log, storage))
	router.Get("/user_banner", get.New(log, storage, redis))

	router.Post("/banner", save.New(log, storage, redis))
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
