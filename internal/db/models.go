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
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	ExternalID int             `json:"external_id"` // TVMaze ID for TV shows, MAL ID for anime
	ShowType   string          `json:"show_type"`   // "tv" or "anime"
	Name       string          `json:"name"`
	Summary    *string         `json:"summary"`
	ImageURL   *string         `json:"image_url"`
	Status     *string         `json:"status"`
	Premiered  *time.Time      `json:"premiered"`
	Ended      *time.Time      `json:"ended"`
	Network    *string         `json:"network"`
	Genres     json.RawMessage `json:"genres"`
	Rating     json.RawMessage `json:"rating"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	
	// Episode counts (not stored in DB, calculated in queries)
	TotalEpisodes   int `json:"total_episodes"`
	WatchedEpisodes int `json:"watched_episodes"`
	
	// Legacy field for backward compatibility - will be removed later
	TVMazeID int `json:"tvmaze_id,omitempty"`
}

// Episode entity
type Episode struct {
	ID         uuid.UUID  `json:"id"`
	ShowID     uuid.UUID  `json:"show_id"`
	UserID     uuid.UUID  `json:"user_id"`
	ExternalID int        `json:"external_id"` // TVMaze ID for TV shows, MAL ID for anime
	ShowType   string     `json:"show_type"`   // "tv" or "anime"
	Name       string     `json:"name"`
	Season     int        `json:"season"`
	Number     int        `json:"number"`
	Summary    *string    `json:"summary"`
	AirDate    *time.Time `json:"airdate"`
	Runtime    *int       `json:"runtime"`
	ImageURL   *string    `json:"image_url"`
	Filler     bool       `json:"filler"`      // For anime episodes - indicates filler content
	Recap      bool       `json:"recap"`       // For anime episodes - indicates recap content
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	
	// Legacy field for backward compatibility - will be removed later
	TVMazeID int `json:"tvmaze_id,omitempty"`
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

// Combined search result that can hold both TV shows and anime
type UnifiedSearchResult struct {
	Type         string                 `json:"type"`         // "tv" or "anime"
	TVShow       *TVMazeSearchResult    `json:"tv_show,omitempty"`
	Anime        *JikanAnime           `json:"anime,omitempty"`
	AlreadyAdded bool                   `json:"already_added"`
}

// Jikan API Response structs (moved from api package for better organization)
type JikanAnime struct {
	MalID         int             `json:"mal_id"`
	URL           string          `json:"url"`
	Images        JikanImages     `json:"images"`
	Title         string          `json:"title"`
	TitleEnglish  *string         `json:"title_english"`
	TitleJapanese *string         `json:"title_japanese"`
	Type          string          `json:"type"`
	Source        string          `json:"source"`
	Episodes      *int            `json:"episodes"`
	Status        string          `json:"status"`
	Airing        bool            `json:"airing"`
	Synopsis      *string         `json:"synopsis"`
	Score         *float64        `json:"score"`
	Year          *int            `json:"year"`
	Studios       []JikanMalItem  `json:"studios"`
	Genres        []JikanMalItem  `json:"genres"`
}

type JikanImages struct {
	JPG  JikanImageFormat `json:"jpg"`
	WebP JikanImageFormat `json:"webp"`
}

type JikanImageFormat struct {
	ImageURL      string `json:"image_url"`
	SmallImageURL string `json:"small_image_url"`
	LargeImageURL string `json:"large_image_url"`
}

type JikanMalItem struct {
	MalID int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// Jikan Episode struct
type JikanEpisode struct {
	MalID         int     `json:"mal_id"`
	URL           *string `json:"url"`
	Title         string  `json:"title"`
	TitleJapanese *string `json:"title_japanese"`
	TitleRomanji  *string `json:"title_romanji"`
	Aired         *string `json:"aired"`
	Score         *float64 `json:"score"`
	Filler        bool    `json:"filler"`
	Recap         bool    `json:"recap"`
	ForumURL      *string `json:"forum_url"`
}

// Jikan Episode Videos struct
type JikanEpisodeVideo struct {
	MalID   int    `json:"mal_id"`
	Title   string `json:"title"`
	Episode string `json:"episode"`
	URL     string `json:"url"`
	Images  struct {
		JPG struct {
			ImageURL string `json:"image_url"`
		} `json:"jpg"`
	} `json:"images"`
}

// Project Management Models

// Project entity
type Project struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Calculated fields (not stored in DB)
	MainTasks      int `json:"main_tasks"`
	SubTasks       int `json:"sub_tasks"`
	CompletedTasks int `json:"completed_tasks"`
}

// Task entity
type Task struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	ProjectID    uuid.UUID  `json:"project_id"`
	ParentTaskID *uuid.UUID `json:"parent_task_id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	Status       string     `json:"status"`
	Priority     string     `json:"priority"`
	DueDate      *time.Time `json:"due_date"`
	Completed    bool       `json:"completed"`
	Collapsed    bool       `json:"collapsed"`
	DisplayOrder int        `json:"display_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	
	// Related data (not stored in DB)
	Subtasks     []Task           `json:"subtasks,omitempty"`
	Dependencies []TaskDependency `json:"dependencies,omitempty"`
	Comments     []TaskComment    `json:"comments,omitempty"`
	Attachments  []TaskAttachment `json:"attachments,omitempty"`
	IsBlocked    bool             `json:"is_blocked"`
}

// Task dependency entity
type TaskDependency struct {
	ID              uuid.UUID `json:"id"`
	TaskID          uuid.UUID `json:"task_id"`
	DependsOnTaskID uuid.UUID `json:"depends_on_task_id"`
	CreatedAt       time.Time `json:"created_at"`
	
	// Related data
	DependsOnTask *Task `json:"depends_on_task,omitempty"`
}

// Task comment entity
type TaskComment struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Task attachment entity
type TaskAttachment struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Filename  string    `json:"filename"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

// Task hierarchy with nested structure for templates
type TaskHierarchy struct {
	Task
	Level       int             `json:"level"`
	HasChildren bool            `json:"has_children"`
	Children    []TaskHierarchy `json:"children,omitempty"`
}

// Market Data Models

// MarketData represents current market data for crypto/stocks
type MarketData struct {
	ID            uuid.UUID `json:"id"`
	Symbol        string    `json:"symbol"`
	CurrentPrice  float64   `json:"current_price"`
	ChangeAmount  float64   `json:"change_amount"`
	ChangePercent float64   `json:"change_percent"`
	Volume        int64     `json:"volume"`
	MarketCap     int64     `json:"market_cap"`
	LastUpdated   time.Time `json:"last_updated"`
}

// MarketWatchlist represents user's preferred market symbols to track
type MarketWatchlist struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Symbol       string    `json:"symbol"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // "crypto", "stock", "forex"
	DisplayOrder int       `json:"display_order"`
	Visible      bool      `json:"visible"`
	CreatedAt    time.Time `json:"created_at"`
}
