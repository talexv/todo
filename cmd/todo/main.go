package main

import (
	"errors"
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

//nolint:sloglint, noctx // will be considered later
func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func setupRouter(db *task.DB) http.Handler {
	router := http.NewServeMux()
	task.NewHandler(router, db)

	return middleware.Recoverer(middleware.Logging(router))
}

//nolint:sloglint, noctx // will be considered later
func run() error {
	if err := godotenv.Load(); err != nil {
		slog.Warn("файл .env не найден",
			"error", err,
		)
	}

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return errors.New("переменная DATABASE_URL не найдена")
	}

	db, err := task.NewDB(connString)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	defer db.Close()

	server := http.Server{
		Addr:              ":8081",
		Handler:           setupRouter(db),
		ReadHeaderTimeout: DefaultTimeout,
	}

	slog.Info("HTTP сервер запущен",
		"url", "http://localhost:8081",
	)

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("ошибка запуска HTTP сервера: %w", err)
	}

	return nil
}
