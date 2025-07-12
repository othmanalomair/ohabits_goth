package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DailyNote represents the aggregated note for one day.
type DailyNote struct {
	Day   int     `json:"day"`
	Note  string  `json:"note"`
	Todos []Todos `json:"todos"`
}

func GetNoteByDate(db *pgxpool.Pool, dateStr string, userID uuid.UUID) (Notes, error) {
	var note Notes
	err := DB.QueryRow(context.Background(), `
        SELECT id, user_id, date, text, created_at, updated_at
        FROM notes
        WHERE date = $1 AND user_id = $2
    `, dateStr, userID).Scan(&note.ID, &note.UserID, &note.Date, &note.Text, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Parse the dateStr (if it fails, use now)
			d, perr := time.Parse("2006-01-02", dateStr)
			if perr != nil {
				d = time.Now()
			}
			// Return an empty note with the provided date (ID remains uuid.Nil)
			return Notes{Date: d}, nil
		}
		return Notes{}, err
	}
	return note, nil
}
func CreateNote(db *pgxpool.Pool, note Notes, userID uuid.UUID) error {
	_, err := DB.Exec(context.Background(), "INSERT INTO notes (user_id, date, text) VALUES ($1, $2, $3)", userID, note.Date, note.Text)
	if err != nil {
		return err
	}
	return nil
}

func UpdateNote(db *pgxpool.Pool, note Notes, noteID uuid.UUID, userID uuid.UUID) error {
	_, err := DB.Exec(context.Background(), "UPDATE notes SET text = $1 WHERE id = $2 AND user_id = $3", note.Text, noteID, userID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteNote(db *pgxpool.Pool, noteID uuid.UUID, userID uuid.UUID) error {
	_, err := DB.Exec(context.Background(), "DELETE FROM notes WHERE id = $1 AND user_id = $2", noteID, userID)
	if err != nil {
		return err
	}
	return nil
}

func GetNotesByMonth(db *pgxpool.Pool, month string, userID uuid.UUID) ([]DailyNote, error) {
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

	var notes []DailyNote
	for day := 1; day <= daysInMonth; day++ {
		currentDate := time.Date(year, mon, day, 0, 0, 0, 0, loc)
		var note string
		err := DB.QueryRow(context.Background(), "SELECT text FROM notes WHERE user_id = $1 AND date = $2", userID, currentDate).Scan(&note)
		if err != nil {
			if err == pgx.ErrNoRows {
				note = ""
			} else {
				return nil, err
			}
		}

		// Get todos for this date
		dateStr := currentDate.Format("2006-01-02")
		todos, err := GetTodosByDate(db, dateStr, userID)
		if err != nil {
			// If there's an error getting todos, continue with empty todos slice
			todos = []Todos{}
		}

		notes = append(notes, DailyNote{Day: day, Note: note, Todos: todos})
	}
	return notes, nil
}
