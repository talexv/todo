package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/talexv/todo/internal/middleware"
	"github.com/talexv/todo/internal/task"
)

const DefaultTimeout = 5 * time.Second

func main() {
	if err := run(); err != nil {
		//nolint:sloglint // will be consired
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("ошибка загрузки .env файла: %w", err)
	}

	db, err := task.NewDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	defer db.Close()

	router := http.NewServeMux()
	task.NewHandler(router, db)

	server := http.Server{
		Addr:              ":8081",
		Handler:           middleware.Recoverer(middleware.Logging(router)),
		ReadHeaderTimeout: DefaultTimeout,
	}

	//nolint:sloglint // will be consired
	slog.Info("HTTP сервер запущен",
		"url", "http://localhost:8081",
	)

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("ошибка запуска HTTP сервера: %w", err)
	}

	return nil
}
