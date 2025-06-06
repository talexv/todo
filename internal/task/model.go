package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

const schemaSQL = `
	CREATE TABLE IF NOT EXISTS tasks (
		id bigserial primary key,
		title text NOT NULL,
		done BOOLEAN default false
	);
`

const getAllTasksSQL = `
	SELECT id, title, done FROM tasks;
`

const createTaskSQL = `
	INSERT INTO tasks (title)
	VALUES ($1)
	RETURNING id, title, done;
`

const updateStatusTaskSQL = `
	UPDATE tasks
	SET done = true
	WHERE id = $1
	RETURNING id, title, done;
`

const deleteTaskSQL = `
	DELETE FROM tasks WHERE id = $1
`

var ErrTaskNotFound = errors.New("task not found")

type DB struct {
	conn *pgx.Conn
}

func NewDB(connString string) (*DB, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("pgx.Connect: %w", err)
	}

	_, err = conn.Exec(context.Background(), schemaSQL)
	if err != nil {
		return nil, fmt.Errorf("conn.Exec: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (d *DB) Close() error {
	err := d.conn.Close(context.Background())
	d.conn = nil

	return err
}

func (d *DB) InsertTask(ctx context.Context, title string) (*Task, error) {
	var task Task

	err := d.conn.QueryRow(ctx, createTaskSQL, title).Scan(&task.ID, &task.Title, &task.Done)
	if err != nil {
		return nil, fmt.Errorf("conn.QueryRow: %w", err)
	}

	return &task, nil
}

func (d *DB) GetAllTasks(ctx context.Context) ([]*Task, error) {
	rows, err := d.conn.Query(ctx, getAllTasksSQL)
	if err != nil {
		return nil, fmt.Errorf("conn.Query: %w", err)
	}

	defer rows.Close()

	var tasks []*Task

	for rows.Next() {
		task := &Task{}

		err = rows.Scan(&task.ID, &task.Title, &task.Done)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		tasks = append(tasks, task)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return tasks, nil
}

func (d *DB) DeleteTask(ctx context.Context, id int64) error {
	tag, err := d.conn.Exec(ctx, deleteTaskSQL, id)
	if err != nil {
		return fmt.Errorf("conn.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (d *DB) UpdateStatusTask(ctx context.Context, id int64) (*Task, error) {
	var task Task

	err := d.conn.QueryRow(ctx, updateStatusTaskSQL, id).Scan(&task.ID, &task.Title, &task.Done)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}

		return nil, fmt.Errorf("conn.QueryRow: %w", err)
	}

	return &task, nil
}
