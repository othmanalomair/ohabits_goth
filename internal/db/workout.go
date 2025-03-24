package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAllWorkouts(db *pgxpool.Pool, userID uuid.UUID) ([]Workout, error) {
	// Get all workouts from the postgres database
	workouts := []Workout{}

	rows, err := DB.Query(context.Background(), "SELECT * FROM workouts WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var workout Workout
		err := rows.Scan(&workout.ID, &workout.UserID, &workout.Name, &workout.Day, &workout.Exercises, &workout.CreatedAt, &workout.UpdatedAt)
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

	err := DB.QueryRow(context.Background(), "SELECT * FROM workouts WHERE id = $1 AND user_id = $2", workoutID, userID).Scan(&workout.ID, &workout.UserID, &workout.Name, &workout.Day, &workout.Exercises, &workout.CreatedAt, &workout.UpdatedAt)
	if err != nil {
		return Workout{}, err
	}

	return workout, nil
}

func CreateWorkout(db *pgxpool.Pool, workout Workout, userID uuid.UUID) error {
	// Create a new workout in the postgres database
	_, err := DB.Exec(
		context.Background(), `
		INSERT INTO workouts (user_id, name, day, exercises)
		VALUES ($1, $2, $3, $4)
		`, userID, workout.Name, workout.Day, workout.Exercises)
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

func GetWorkoutLogByDate(db *pgxpool.Pool, logDate time.Time, userID uuid.UUID) (WorkoutLog, error) {
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
	// Create a new workout log in the postgres database
	_, err := DB.Exec(
		context.Background(), `
		INSERT INTO workout_logs (user_id, name, completed_exercises, cardio, weight, date)
		VALUES ($1, $2, $3, $4, $5, $6)
		`, userID, workoutLog.Name, workoutLog.CompletedExercises, workoutLog.Cardio, workoutLog.Weight, workoutDate)
	if err != nil {
		return err
	}

	return nil
}
