package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Data      *string    `json:"data,omitempty"` // JSON data as string
	Read      bool       `json:"read"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CreateNotification creates a new notification
func CreateNotification(db *pgxpool.Pool, userID uuid.UUID, notificationType, title, message string, data *string) error {
	_, err := db.Exec(context.Background(), `
		INSERT INTO notifications (user_id, type, title, message, data, read, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, false, NOW(), NOW())`,
		userID, notificationType, title, message, data)
	return err
}

// GetUnreadNotificationCount returns the count of unread notifications for a user
func GetUnreadNotificationCount(db *pgxpool.Pool, userID uuid.UUID) (int, error) {
	var count int
	err := db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM notifications 
		WHERE user_id = $1 AND read = false`,
		userID).Scan(&count)
	return count, err
}

// GetNotifications returns paginated unread notifications for a user
func GetNotifications(db *pgxpool.Pool, userID uuid.UUID, limit, offset int) ([]Notification, error) {
	rows, err := db.Query(context.Background(), `
		SELECT id, user_id, type, title, message, data, read, created_at, updated_at
		FROM notifications 
		WHERE user_id = $1 AND read = false
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var notification Notification
		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.Type, 
			&notification.Title, &notification.Message, &notification.Data,
			&notification.Read, &notification.CreatedAt, &notification.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, rows.Err()
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(db *pgxpool.Pool, notificationID, userID uuid.UUID) error {
	_, err := db.Exec(context.Background(), `
		UPDATE notifications 
		SET read = true, updated_at = NOW() 
		WHERE id = $1 AND user_id = $2`,
		notificationID, userID)
	return err
}

// DeleteNotification deletes a notification
func DeleteNotification(db *pgxpool.Pool, notificationID, userID uuid.UUID) error {
	_, err := db.Exec(context.Background(), `
		DELETE FROM notifications 
		WHERE id = $1 AND user_id = $2`,
		notificationID, userID)
	return err
}

// CreateEpisodeNotification creates a notification for new episodes
func CreateEpisodeNotification(db *pgxpool.Pool, userID uuid.UUID, showName string, episodeCount int, showID uuid.UUID) error {
	var title, message string
	var data *string
	
	if episodeCount == 1 {
		title = "New Episode Available!"
		message = "A new episode of " + showName + " is now available."
	} else {
		title = "New Episodes Available!"
		message = fmt.Sprintf("%d new episodes of %s are now available.", episodeCount, showName)
	}
	
	// Create JSON data with show information
	if showID != uuid.Nil {
		jsonData := fmt.Sprintf(`{"show_id": "%s", "episode_count": %d}`, showID.String(), episodeCount)
		data = &jsonData
	}
	
	return CreateNotification(db, userID, "new_episodes", title, message, data)
}

// CleanupOldNotifications removes notifications older than 30 days
func CleanupOldNotifications(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), `
		DELETE FROM notifications 
		WHERE created_at < NOW() - INTERVAL '30 days'`)
	return err
}