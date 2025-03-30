package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetRateByDate(db *pgxpool.Pool, date string, userID uuid.UUID) (int, error) {
	var rate int

	err := db.QueryRow(context.Background(), "SELECT rating FROM mood_ratings WHERE date = $1 AND user_id = $2", date, userID).Scan(&rate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return rate, nil
}

func GetMoodRatingByDate(db *pgxpool.Pool, dateStr string, userID uuid.UUID) (MoodRating, error) {
	var mr MoodRating
	err := db.QueryRow(context.Background(), `
        SELECT id, user_id, rating, date, created_at, updated_at
        FROM mood_ratings
        WHERE date::date = to_date($1, 'YYYY-MM-DD') AND user_id = $2
    `, dateStr, userID).Scan(&mr.ID, &mr.UserID, &mr.Rating, &mr.Date, &mr.CreatedAt, &mr.UpdatedAt)
	if err != nil {
		return MoodRating{}, err
	}
	return mr, nil
}

func CreateRate(db *pgxpool.Pool, rate MoodRating, userID uuid.UUID) error {
	_, err := db.Exec(context.Background(), "INSERT INTO mood_ratings (date, user_id, rating) VALUES ($1, $2, $3)", rate.Date, userID, rate.Rating)
	if err != nil {
		return err
	}
	return nil
}

func UpdateRate(db *pgxpool.Pool, rate MoodRating, rateID uuid.UUID, userID uuid.UUID) error {
	_, err := db.Exec(context.Background(), "UPDATE mood_ratings SET rating = $1 WHERE id = $2 AND user_id = $3", rate.Rating, rateID, userID)
	if err != nil {
		return err
	}
	return nil
}
