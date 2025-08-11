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
	AvatarURL   *string   `json:"avatar_url"`
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
type Exercise struct {
	Order int    `json:"order"`
	Name  string `json:"name"`
}

type Workout struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Name         string     `json:"name"`
	Day          string     `json:"day"`
	Exercises    []Exercise `json:"exercises"`
	DisplayOrder int        `json:"display_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
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

// Medication entity
type Medication struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	Name          string          `json:"name"`
	Dosage        string          `json:"dosage"`
	ScheduledDays json.RawMessage `json:"scheduled_days"`
	TimesPerDay   int             `json:"times_per_day"`
	TimeIntervals []string        `json:"time_intervals"`
	DurationType  string          `json:"duration_type"`
	StartDate     time.Time       `json:"start_date"`
	EndDate       *time.Time      `json:"end_date"`
	Notes         *string         `json:"notes"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Medication Log entity
type MedicationLog struct {
	ID            uuid.UUID  `json:"id"`
	MedicationID  uuid.UUID  `json:"medication_id"`
	UserID        uuid.UUID  `json:"user_id"`
	Taken         bool       `json:"taken"`
	ScheduledTime string     `json:"scheduled_time"`
	ActualTime    *time.Time `json:"actual_time"`
	Date          time.Time  `json:"date"`
	Notes         *string    `json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Show entity
type Show struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	TVMazeID  int             `json:"tvmaze_id"`
	Name      string          `json:"name"`
	Summary   *string         `json:"summary"`
	ImageURL  *string         `json:"image_url"`
	Status    *string         `json:"status"`
	Premiered *time.Time      `json:"premiered"`
	Ended     *time.Time      `json:"ended"`
	Network   *string         `json:"network"`
	Genres    json.RawMessage `json:"genres"`
	Rating    json.RawMessage `json:"rating"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Episode entity
type Episode struct {
	ID       uuid.UUID  `json:"id"`
	ShowID   uuid.UUID  `json:"show_id"`
	UserID   uuid.UUID  `json:"user_id"`
	TVMazeID int        `json:"tvmaze_id"`
	Name     string     `json:"name"`
	Season   int        `json:"season"`
	Number   int        `json:"number"`
	Summary  *string    `json:"summary"`
	AirDate  *time.Time `json:"airdate"`
	Runtime  *int       `json:"runtime"`
	ImageURL *string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Episode Tracking entity
type EpisodeTracking struct {
	ID        uuid.UUID  `json:"id"`
	EpisodeID uuid.UUID  `json:"episode_id"`
	UserID    uuid.UUID  `json:"user_id"`
	Watched   bool       `json:"watched"`
	Rating    *int       `json:"rating"`
	Notes     *string    `json:"notes"`
	WatchedAt *time.Time `json:"watched_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Combined Episode with Tracking data for templates
type EpisodeWithTracking struct {
	Episode
	Watched   bool       `json:"watched"`
	Rating    *int       `json:"rating"`
	Notes     *string    `json:"notes"`
	WatchedAt *time.Time `json:"watched_at"`
}

// TVMaze API Response structs
type TVMazeSearchResult struct {
	Score float64      `json:"score"`
	Show  TVMazeShow   `json:"show"`
}

type TVMazeShow struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Summary    *string           `json:"summary"`
	Image      *TVMazeImage      `json:"image"`
	Status     string            `json:"status"`
	Premiered  *string           `json:"premiered"`
	Ended      *string           `json:"ended"`
	Network    *TVMazeNetwork    `json:"network"`
	WebChannel *TVMazeNetwork    `json:"webChannel"`
	Genres     []string          `json:"genres"`
	Rating     *TVMazeRating     `json:"rating"`
}

type TVMazeImage struct {
	Medium   *string `json:"medium"`
	Original *string `json:"original"`
}

type TVMazeNetwork struct {
	Name string `json:"name"`
}

type TVMazeRating struct {
	Average *float64 `json:"average"`
}

type TVMazeEpisode struct {
	ID      int          `json:"id"`
	Name    string       `json:"name"`
	Season  int          `json:"season"`
	Number  int          `json:"number"`
	Summary *string      `json:"summary"`
	AirDate *string      `json:"airdate"`
	Runtime *int         `json:"runtime"`
	Image   *TVMazeImage `json:"image"`
}
