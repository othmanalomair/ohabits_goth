package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAllHabits(db *pgxpool.Pool, userID uuid.UUID) ([]Habit, error) {
	// Get all habits from the postgres database
	habits := []Habit{}

	rows, err := db.Query(context.Background(), "SELECT id, user_id, name, scheduled_days, created_at, updated_at FROM habits WHERE user_id = $1", userID)
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

func CreateHabit(db *pgxpool.Pool, habit Habit, userID uuid.UUID) error {
	// Create a habit in the postgres database
	_, err := db.Exec(context.Background(), `
		INSERT INTO habits (user_id, name, scheduled_days)
		VALUES ($1, $2, $3)
		`, userID, habit.Name, habit.ScheduledDays)
	return err
}

func UpdateHabit(db *pgxpool.Pool, habit Habit, userID uuid.UUID) error {
	// Update a habit in the postgres database
	_, err := db.Exec(context.Background(), `
		UPDATE habits SET name = $2, scheduled_days = $3 WHERE id = $1 AND user_id = $4
		`, habit.ID, habit.Name, habit.ScheduledDays, userID)
	return err
}

func DeleteHabit(db *pgxpool.Pool, habit Habit) error {
	// Delete a habit in the postgres database
	_, err := db.Exec(context.Background(), `
		DELETE FROM habits WHERE id = $1
		`, habit.ID)
	return err
}

func GetHabitsCompletedByDate(db *pgxpool.Pool, date string, userID uuid.UUID) ([]HabitCompletion, error) {
	// Get habits by date from the postgres database
	habits := []HabitCompletion{}

	rows, err := db.Query(context.Background(), `
		SELECT hc.id, hc.habit_id, h.name, hc.user_id, hc.completed, hc.date
		FROM habits_completions hc
		JOIN habits h ON hc.habit_id = h.id
		WHERE hc.date = $1 AND hc.user_id = $2
		`, date, userID)
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

func CreateHabitCompletion(db *pgxpool.Pool, habitCompletion HabitCompletion, userID uuid.UUID) error {
	// Create a habit completion in the postgres database
	_, err := db.Exec(context.Background(), `
		INSERT INTO habits_completions (habit_id, user_id, completed, date)
		VALUES ($1, $2, $3, $4)
		`, habitCompletion.HabitID, userID, false, habitCompletion.Date)
	return err
}

func UpdateHabitCompletion(db *pgxpool.Pool, habitCompletion HabitCompletion, userID uuid.UUID) error {
	// Update a habit completion in the postgres database
	_, err := db.Exec(context.Background(), `
		UPDATE habits_completions SET completed = $1 WHERE id = $2 AND user_id = $3
		`, habitCompletion.Completed, habitCompletion.ID, userID)
	return err
}
