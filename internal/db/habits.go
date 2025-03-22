package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAllHabits(db *pgxpool.Pool) ([]Habit, error) {
	// Get all habits from the postgres database
	habits := []Habit{}

	rows, err := db.Query(context.Background(), "SELECT id, user_id, name, scheduled_days, created_at, updated_at FROM habits")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var habit Habit
		err := rows.Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.ScheduledDays, &habit.CreatedAt, &habit.UpdatedAt)
		if err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}
	return habits, nil
}

func GetHabitsCompletedByDate(db *pgxpool.Pool, date string) ([]HabitCompletion, error) {
	// Get habits by date from the postgres database
	habits := []HabitCompletion{}

	rows, err := db.Query(context.Background(), `
		SELECT hc.id, hc.habit_id, h.name, hc.user_id, hc.completed, hc.date
		FROM habits_completions hc
		JOIN habits h ON hc.habit_id = h.id
		WHERE hc.date = $1
		`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var habit HabitCompletion
		err := rows.Scan(&habit.ID, &habit.HabitID, &habit.HabitName, &habit.UserID, &habit.Completed, &habit.Date)
		if err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}
	return habits, nil
}
