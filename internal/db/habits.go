package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAllHabits(db *pgxpool.Pool) ([]Habits, error) {
	// Get all habits from the postgres database
	habits := []Habits{}

	rows, err := db.Query(context.Background(), "SELECT id, user_id, name, scheduled_days, created_at, updated_at FROM habits")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var habit Habits
		err := rows.Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.ScheduledDays, &habit.CreatedAt, &habit.UpdatedAt)
		if err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}
	return habits, nil
}
