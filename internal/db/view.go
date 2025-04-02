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
		// Retrieve habits that are scheduled for today.
		var habits []struct {
			ID   uuid.UUID
			Name string
			Days []string
		}
		rows, err := db.Query(context.Background(), `
    SELECT id, name, scheduled_days
    FROM habits
    WHERE user_id = $1
`, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var h struct {
				ID   uuid.UUID
				Name string
				Days []string
			}
			if err := rows.Scan(&h.ID, &h.Name, &h.Days); err != nil {
				return nil, err
			}
			habits = append(habits, h)
		}

		// Count how many of today's habits are completed
		var totalHabits, completedHabits int
		for _, habit := range habits {
			for _, day := range habit.Days {
				// Check if the current day matches one of the scheduled days
				if day == currentDate.Weekday().String() {
					totalHabits++
					// Check if the habit was completed today
					var completed bool
					err = db.QueryRow(context.Background(), `
                        SELECT completed
                        FROM habits_completions
                        WHERE user_id = $1 AND habit_id = $2 AND date = $3
                    `, userID, habit.ID, currentDate).Scan(&completed)
					if err == nil && completed {
						completedHabits++
					}
					break
				}
			}
		}

		daily.View.Habits = fmt.Sprintf("%d/%d", completedHabits, totalHabits)

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
