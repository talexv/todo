package task

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type Handler struct {
	db *DB
}

func NewHandler(router *http.ServeMux, db *DB) {
	handler := &Handler{db: db}

	router.HandleFunc("GET /tasks", handler.GetTasks())
	router.HandleFunc("POST /create", handler.CreateTask())
	router.HandleFunc("PATCH /tasks/{id}/done", handler.UpdateStatusTask())
	router.HandleFunc("DELETE /tasks/{id}/delete", handler.DeleteTask())
	// router.HandleFunc("GET /panic", handler.TestPanic())
}

func (handler *Handler) GetTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := handler.db.GetAllTasks(r.Context())
		if err != nil {
			http.Error(w, "ошибка при получении списка задач", http.StatusInternalServerError)
			return
		}

		writeJSON(w, tasks, http.StatusOK)
	}
}

func (handler *Handler) CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "некорректный JSON", http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			http.Error(w, "'title' обязательно к заполнению", http.StatusBadRequest)
			return
		}

		task, err := handler.db.InsertTask(r.Context(), req.Title)
		if err != nil {
			http.Error(w, "ошибка при создании задачи", http.StatusInternalServerError)
			return
		}

		writeJSON(w, task, http.StatusCreated)
	}
}

func (handler *Handler) DeleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = handler.db.DeleteTask(r.Context(), id)
		if err != nil {
			if errors.Is(err, ErrTaskNotFound) {
				http.Error(w, "задача не найдена", http.StatusNotFound)
				return
			}

			http.Error(w, "ошибка при удалении задачи", http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (handler *Handler) UpdateStatusTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		task, err := handler.db.UpdateStatusTask(r.Context(), id)
		if err != nil {
			if errors.Is(err, ErrTaskNotFound) {
				http.Error(w, "задача не найдена", http.StatusNotFound)
				return
			}

			http.Error(w, "ошибка при обновлении статуса задачи", http.StatusInternalServerError)

			return
		}

		writeJSON(w, task, http.StatusOK)
	}
}

func writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func parseID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil || id <= 0 {
		return 0, errors.New("некорректный id")
	}

	return id, nil
}

// func (handler *Handler) TestPanic() http.HandlerFunc {
// 	return func(http.ResponseWriter, *http.Request) {
// 		panic("паника")
// 	}
// }
