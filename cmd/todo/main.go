package main

import (
	"fmt"
	"log"
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
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("ошибка загрузки .env файла: %w", err)
	}

	conn, err := task.NewDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("ошибка при закрытии соединения c БД: %v", closeErr)
		}
	}()

	router := http.NewServeMux()
	task.NewHandler(router, conn)

	server := http.Server{
		Addr:              ":8081",
		Handler:           middleware.Recoverer(middleware.Logging(router)),
		ReadHeaderTimeout: DefaultTimeout,
	}

	log.Println("HTTP сервер запущен на http://localhost:8081")

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("ошибка запуска HTTP сервера: %w", err)
	}

	return nil
}
