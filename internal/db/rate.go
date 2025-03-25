package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetRateByDate(db *pgxpool.Pool, date string, userID uuid.UUID) (int, error) {
	var rate int

	err := db.QueryRow(context.Background(), "SELECT rating FROM mood_ratings WHERE date = $1 AND user_id = $2", date, userID).Scan(&rate)
	if err != nil {
		return 0, err
	}
	return rate, nil
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
