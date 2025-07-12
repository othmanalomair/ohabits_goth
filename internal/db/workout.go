package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAllWorkouts(db *pgxpool.Pool, userID uuid.UUID) ([]Workout, error) {
	// Get all workouts from the postgres database ordered by display_order
	workouts := []Workout{}

	rows, err := DB.Query(context.Background(), "SELECT id, user_id, name, day, exercises, display_order, created_at, updated_at FROM workouts WHERE user_id = $1 ORDER BY display_order", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var workout Workout
		err := rows.Scan(&workout.ID, &workout.UserID, &workout.Name, &workout.Day, &workout.Exercises, &workout.DisplayOrder, &workout.CreatedAt, &workout.UpdatedAt)
		if err != nil {
			return nil, err
		}
		workouts = append(workouts, workout)
	}

	return workouts, nil
}

func GetWorkout(db *pgxpool.Pool, workoutID uuid.UUID, userID uuid.UUID) (Workout, error) {
	// Get a workout from the postgres database
	var workout Workout

	err := DB.QueryRow(context.Background(), "SELECT id, user_id, name, day, exercises, display_order, created_at, updated_at FROM workouts WHERE id = $1 AND user_id = $2", workoutID, userID).Scan(&workout.ID, &workout.UserID, &workout.Name, &workout.Day, &workout.Exercises, &workout.DisplayOrder, &workout.CreatedAt, &workout.UpdatedAt)
	if err != nil {
		return Workout{}, err
	}

	return workout, nil
}

func CreateWorkout(db *pgxpool.Pool, workout Workout, userID uuid.UUID) error {
	// Create a new workout in the postgres database with the next display_order
	var nextOrder int
	err := DB.QueryRow(context.Background(), "SELECT COALESCE(MAX(display_order), 0) + 1 FROM workouts WHERE user_id = $1", userID).Scan(&nextOrder)
	if err != nil {
		return err
	}

	_, err = DB.Exec(
		context.Background(), `
		INSERT INTO workouts (user_id, name, day, exercises, display_order)
		VALUES ($1, $2, $3, $4, $5)
		`, userID, workout.Name, workout.Day, workout.Exercises, nextOrder)
	if err != nil {
		return err
	}

	return nil
}

func UpdateWorkout(db *pgxpool.Pool, workoutID uuid.UUID, workout Workout, userID uuid.UUID) error {
	// Update a workout in the postgres database
	_, err := DB.Exec(
		context.Background(), `
		UPDATE workouts SET name = $1, day = $2, exercises = $3 WHERE id = $4 AND user_id = $5
		`, workout.Name, workout.Day, workout.Exercises, workoutID, userID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteWorkout(db *pgxpool.Pool, workoutID uuid.UUID, userID uuid.UUID) error {
	// Delete a workout from the postgres database
	_, err := DB.Exec(
		context.Background(), `
		DELETE FROM workouts WHERE id = $1 AND user_id = $2
		`, workoutID, userID)
	if err != nil {
		return err
	}

	return nil
}

func GetWorkoutLogByDate(db *pgxpool.Pool, logDate string, userID uuid.UUID) (WorkoutLog, error) {
	var log WorkoutLog

	err := db.QueryRow(
		context.Background(),
		`SELECT id, user_id, name, completed_exercises, cardio, weight, date, created_at, updated_at
		 FROM workout_logs
		 WHERE date = $1 AND user_id = $2`,
		logDate, userID,
	).Scan(
		&log.ID,
		&log.UserID,
		&log.Name,
		&log.CompletedExercises,
		&log.Cardio,
		&log.Weight,
		&log.Date,
		&log.CreatedAt,
		&log.UpdatedAt,
	)

	if err != nil {
		return WorkoutLog{}, err
	}

	return log, nil
}

func CreateWorkoutLog(db *pgxpool.Pool, workoutLog WorkoutLog, userID uuid.UUID, workoutDate time.Time) error {
	_, err := DB.Exec(context.Background(), `
        INSERT INTO workout_logs (user_id, name, completed_exercises, cardio, weight, date)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, userID, workoutLog.Name, workoutLog.CompletedExercises, workoutLog.Cardio, workoutLog.Weight, workoutDate)
	return err
}

func UpdateWorkoutLog(db *pgxpool.Pool, logID uuid.UUID, logEntry WorkoutLog, userID uuid.UUID) error {
	_, err := db.Exec(context.Background(), `
		UPDATE workout_logs
		SET name = $1, completed_exercises = $2, cardio = $3, weight = $4, date = $5, updated_at = NOW()
		WHERE id = $6 AND user_id = $7
	`, logEntry.Name, logEntry.CompletedExercises, logEntry.Cardio, logEntry.Weight, logEntry.Date, logID, userID)
	return err
}

func MoveWorkoutUp(db *pgxpool.Pool, workoutID uuid.UUID, userID uuid.UUID) error {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Get current workout's display_order
	var currentOrder int
	err = tx.QueryRow(context.Background(), 
		"SELECT display_order FROM workouts WHERE id = $1 AND user_id = $2", 
		workoutID, userID).Scan(&currentOrder)
	if err != nil {
		return err
	}

	// Can't move up if already at the top
	if currentOrder <= 1 {
		return nil
	}

	// Find the workout with the previous order
	_, err = tx.Exec(context.Background(), `
		UPDATE workouts SET display_order = $1, updated_at = NOW()
		WHERE user_id = $2 AND display_order = $3
	`, currentOrder, userID, currentOrder-1)
	if err != nil {
		return err
	}

	// Move current workout up
	_, err = tx.Exec(context.Background(), `
		UPDATE workouts SET display_order = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
	`, currentOrder-1, workoutID, userID)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func MoveWorkoutDown(db *pgxpool.Pool, workoutID uuid.UUID, userID uuid.UUID) error {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Get current workout's display_order and max order
	var currentOrder, maxOrder int
	err = tx.QueryRow(context.Background(), 
		"SELECT display_order FROM workouts WHERE id = $1 AND user_id = $2", 
		workoutID, userID).Scan(&currentOrder)
	if err != nil {
		return err
	}

	err = tx.QueryRow(context.Background(), 
		"SELECT MAX(display_order) FROM workouts WHERE user_id = $1", 
		userID).Scan(&maxOrder)
	if err != nil {
		return err
	}

	// Can't move down if already at the bottom
	if currentOrder >= maxOrder {
		return nil
	}

	// Find the workout with the next order
	_, err = tx.Exec(context.Background(), `
		UPDATE workouts SET display_order = $1, updated_at = NOW()
		WHERE user_id = $2 AND display_order = $3
	`, currentOrder, userID, currentOrder+1)
	if err != nil {
		return err
	}

	// Move current workout down
	_, err = tx.Exec(context.Background(), `
		UPDATE workouts SET display_order = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
	`, currentOrder+1, workoutID, userID)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}
