package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// func GetNoteByDate(db *pgxpool.Pool, dateStr string, userID uuid.UUID) (Notes, error) {
// 	var note Notes
// 	err := DB.QueryRow(context.Background(), "SELECT id, user_id, date, text, created_at, updated_at FROM notes WHERE date = $1 AND user_id = $2", dateStr, userID).
// 		Scan(&note.ID, &note.UserID, &note.Date, &note.Text, &note.CreatedAt, &note.UpdatedAt)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			// Parse the date string into a time.Time value. If parsing fails, fall back to current time.
// 			d, perr := time.Parse("2006-01-02", dateStr)
// 			if perr != nil {
// 				d = time.Now()
// 			}
// 			// Return an empty note with the given date.
// 			return Notes{
// 				Date: d,
// 			}, nil
// 		}
// 		return Notes{}, err
// 	}
// 	return note, nil
// }

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
