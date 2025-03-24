package db

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User entity
type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Habit entity
type Habit struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	Name          string          `json:"name"`
	ScheduledDays json.RawMessage `json:"scheduled_days"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type HabitCompletion struct {
	ID        uuid.UUID `json:"id"`
	HabitID   uuid.UUID `json:"habit_id"`
	HabitName string    `json:"habit_name"`
	UserID    uuid.UUID `json:"user_id"`
	Completed bool      `json:"completed"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Workout entity
type Workout struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Name      string          `json:"name"`
	Day       string          `json:"day"`
	Exercises json.RawMessage `json:"exercises"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type WorkoutLog struct {
	ID                 uuid.UUID       `json:"id"`
	UserID             uuid.UUID       `json:"user_id"`
	Name               string          `json:"name"`
	CompletedExercises json.RawMessage `json:"completed_exercises"`
	Cardio             json.RawMessage `json:"cardio"`
	Weight             float64         `json:"weight"`
	Date               time.Time       `json:"date"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// Todo's entity
type Todos struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Notes entity
type Notes struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Date      time.Time `json:"date"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Mood ratings entity
type MoodRating struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Rating    int       `json:"rating"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
