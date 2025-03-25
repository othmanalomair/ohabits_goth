package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetUser(db *pgxpool.Pool, userID uuid.UUID) (User, error) {
	var user User
	err := db.QueryRow(context.Background(), `
        SELECT id, email, display_name, avatar_url, created_at, updated_at
        FROM users
        WHERE id = $1
    `, userID).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func UpdateUser(db *pgxpool.Pool, user User, userID uuid.UUID) error {
	_, err := db.Exec(context.Background(), `
		UPDATE users SET email = $1, display_name = $2, avatar_url = $3, updated_at = NOW()
		WHERE id = $4
	`, user.Email, user.DisplayName, user.AvatarURL, userID)
	if err != nil {
		return err
	}
	return nil
}
