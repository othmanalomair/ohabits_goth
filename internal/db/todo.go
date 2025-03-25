package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetTodosByDate(db *pgxpool.Pool, todoDate string, userID uuid.UUID) ([]Todos, error) {
	todos := []Todos{}

	rows, err := db.Query(context.Background(), `
		SELECT id, user_id, text, completed, date, created_at, updated_at
		FROM todos
		WHERE date = $1 AND user_id = $2
		ORDER BY created_at DESC
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todos
		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Text,
			&todo.Completed,
			&todo.Date,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func CreateTodo(db *pgxpool.Pool, todo Todos, userID uuid.UUID) error {

	// Create a new todo
	_, err := db.Exec(context.Background(), `
		INSERT INTO todos (user_id, text, completed, date)
		VALUES ($1, $2, $3, $4)
		`, userID, todo.Text, false, todo.Date)
	if err != nil {
		return err
	}
	return nil
}

func UpdateTodo(db *pgxpool.Pool, todo Todos, id uuid.UUID, userID uuid.UUID) error {

	// Update a todo
	_, err := db.Exec(context.Background(), `
		UPDATE todos
		SET text = $1, completed = $2
		WHERE id = $3 AND user_id = $4
		`, todo.Text, todo.Completed, id, userID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteTodo(db *pgxpool.Pool, id uuid.UUID, userID uuid.UUID) error {

	// Delete a todo
	_, err := db.Exec(context.Background(), `
		DELETE FROM todos
		WHERE id = $1 AND user_id = $2
		`, id, userID)
	if err != nil {
		return err
	}
	return nil
}
