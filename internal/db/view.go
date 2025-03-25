package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DailyView represents the aggregated view for one day.
type DailyView struct {
	Day  int       `json:"day"`
	View DailyData `json:"view"`
}

// DailyData contains the details for each day.
type DailyData struct {
	Workout string          `json:"workout,omitempty"` // workout_logss.name
	Cardio  json.RawMessage `json:"cardio,omitempty"`  // workout_logs.cardio
	Weight  float64         `json:"weight,omitempty"`  // workout_logs.weight
	Habits  string          `json:"habits"`            // formatted as "completed/total"
	Mood    *int            `json:"mood,omitempty"`    // mood_rating.rating
}

// GetViewByMonth queries the database and builds a daily view for the month.
// The month parameter should be in "YYYY-MM" format.
func GetViewByMonth(db *pgxpool.Pool, month string, userID uuid.UUID) ([]DailyView, error) {
	// Parse the month string to get the first day of the month.
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, err
	}
	year, mon, _ := t.Date()
	loc := t.Location()
	firstOfMonth := time.Date(year, mon, 1, 0, 0, 0, 0, loc)
	// Determine the number of days in the month.
	nextMonth := firstOfMonth.AddDate(0, 1, 0)
	lastOfMonth := nextMonth.Add(-time.Hour * 24)
	daysInMonth := lastOfMonth.Day()

	var views []DailyView

	// Loop over each day of the month.
	for day := 1; day <= daysInMonth; day++ {
		currentDate := time.Date(year, mon, day, 0, 0, 0, 0, loc)
		daily := DailyView{
			Day: day,
			View: DailyData{
				Habits: "0/0", // default value if no habits are found
			},
		}

		// --- Workout Query ---
		// Query the workout_logs for the current day.
		var workoutName string
		var cardio json.RawMessage
		var weight float64
		err = db.QueryRow(context.Background(), `
			SELECT name, cardio, weight
			FROM workout_logs
			WHERE user_id = $1 AND date = $2
		`, userID, currentDate).Scan(&workoutName, &cardio, &weight)
		if err == nil {
			daily.View.Workout = workoutName
			daily.View.Cardio = cardio
			daily.View.Weight = weight
		} else if err != pgx.ErrNoRows {
			return nil, err
		}

		// --- Habits Query ---
		// Count total habit completions and how many are completed for the current day.
		var totalHabits int
		var completedHabits int
		err = db.QueryRow(context.Background(), `
			SELECT COUNT(*), COALESCE(SUM(CASE WHEN completed THEN 1 ELSE 0 END), 0)
			FROM habits_completions
			WHERE user_id = $1 AND date = $2
		`, userID, currentDate).Scan(&totalHabits, &completedHabits)
		if err == nil {
			daily.View.Habits = fmt.Sprintf("%d/%d", completedHabits, totalHabits)
		} else if err != pgx.ErrNoRows {
			return nil, err
		}

		// --- Mood Query ---
		// Query the mood rating for the current day.
		var moodRating int
		err = db.QueryRow(context.Background(), `
			SELECT rating
			FROM mood_ratings
			WHERE user_id = $1 AND date = $2
		`, userID, currentDate).Scan(&moodRating)
		if err == nil {
			daily.View.Mood = &moodRating
		} else if err != pgx.ErrNoRows {
			return nil, err
		}

		views = append(views, daily)
	}

	return views, nil
}
