package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

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

	fmt.Println(storage)

	// Test create banner
	fmt.Println("start create")
	id, err := storage.CreateBanner(1, []int{1, 2}, `{"qqq1": "q1q1q"}`, true)
	if err != nil {
		slog.Error("failed to create banner: %w", err)
		os.Exit(1)
	}
	_, err = storage.CreateBanner(2, []int{2, 4}, `{"qqq2": "q2q2q"}`, true)
	if err != nil {
		slog.Error("failed to create banner: %w", err)
		os.Exit(1)
	}

	_, err = storage.CreateBanner(3, []int{1, 2}, `{"qqq3": "q3q3q"}`, true)
	if err != nil {
		slog.Error("failed to create banner: %w", err)
		os.Exit(1)
	}
	fmt.Println(id)
	fmt.Println("finish create")
	fmt.Println("------------------------------------")

	fmt.Println("start update...")
	time.Sleep(7 * time.Second)

	// Test update banner
	err = storage.UpdateBanner(id, 1, []int{2, 3, 4}, `{"qqq4": "q4q4"}`, false)
	if err != nil {
		slog.Error("failed to update banner: %w", err)
		os.Exit(1)
	}
	fmt.Println("finish update")
	fmt.Println("------------------------------------")

	// Get banner list (all banners)
	list, err := storage.GetBannerList(0, 0, 10, 5)
	if err != nil {
		slog.Error("failed to get banner list: %w", err)
		os.Exit(1)
	}
	for _, banner := range list {
		fmt.Println(*banner)
	}

	//Get banner (exist)
	// content, err := storage.GetBanner(id, 1)
	// if err != nil {
	// 	slog.Error("failed to get banner id: %w", err)
	// 	os.Exit(1)
	// }
	// fmt.Println(content)

	// fmt.Println("start delete...")
	// time.Sleep(10 * time.Second)

	// err = storage.DeleteBanner(id)
	// if err != nil {
	// 	slog.Error("failed to get banner id: %w", err)
	// 	os.Exit(1)
	// }

	// fmt.Println("finish delete")
	// fmt.Println("------------------------------------")

	//Get banner (not exist)
	// content, err = storage.GetBanner(id, 1)
	// fmt.Println(content)
	// if err != nil {
	// 	slog.Error("failed to get banner id: %w", err)
	// 	os.Exit(1)
	// }
	// fmt.Println(content)

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
