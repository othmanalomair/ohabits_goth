package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetNoteByDate(db *pgxpool.Pool, date string, userID uuid.UUID) (Notes, error) {
	var note Notes

	err := DB.QueryRow(context.Background(), "SELECT * FROM notes WHERE date = $1 AND user_id = $2", date, userID).Scan(&note.ID, &note.UserID, &note.Date, &note.Text, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
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
