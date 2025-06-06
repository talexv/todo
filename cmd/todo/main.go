package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/talexv/todo/internal/task"
)

const DefaultTimeout = 5 * time.Second

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	conn, err := task.NewDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// defer func() {
	// 	if err = conn.Close(); err != nil {
	// 		log.Printf("Ошибка при закрытии соединениия c БД: %v", err)
	// 	}
	// }()

	router := http.NewServeMux()
	task.NewHandler(router, conn)

	server := http.Server{
		Addr:              ":8081",
		Handler:           router,
		ReadHeaderTimeout: DefaultTimeout,
	}

	fmt.Println("HTTP сервер запущен на http://localhost:8081")

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
	}
}
