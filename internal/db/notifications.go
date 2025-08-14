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

// Project Management Notifications

// CreateTaskDeadlineNotification creates notifications for upcoming task deadlines
func CreateTaskDeadlineNotification(db *pgxpool.Pool, userID uuid.UUID, taskTitle string, projectName string, dueDate time.Time, taskID uuid.UUID) error {
	var title, message string
	var data *string
	
	hoursUntilDue := time.Until(dueDate).Hours()
	
	if hoursUntilDue <= 24 {
		if hoursUntilDue <= 1 {
			title = "Task Due Soon!"
			message = fmt.Sprintf("Task '%s' in project '%s' is due in less than an hour!", taskTitle, projectName)
		} else {
			title = "Task Due Today!"
			message = fmt.Sprintf("Task '%s' in project '%s' is due today at %s", taskTitle, projectName, dueDate.Format("3:04 PM"))
		}
	} else {
		title = "Upcoming Task Deadline"
		message = fmt.Sprintf("Task '%s' in project '%s' is due on %s", taskTitle, projectName, dueDate.Format("Jan 2, 2006"))
	}
	
	// Create JSON data with task information
	if taskID != uuid.Nil {
		jsonData := fmt.Sprintf(`{"task_id": "%s", "due_date": "%s"}`, taskID.String(), dueDate.Format(time.RFC3339))
		data = &jsonData
	}
	
	return CreateNotification(db, userID, "task_deadline", title, message, data)
}

// CreateTaskOverdueNotification creates notifications for overdue tasks
func CreateTaskOverdueNotification(db *pgxpool.Pool, userID uuid.UUID, taskTitle string, projectName string, dueDate time.Time, taskID uuid.UUID) error {
	var title, message string
	var data *string
	
	daysPastDue := int(time.Since(dueDate).Hours() / 24)
	
	title = "Overdue Task"
	if daysPastDue == 0 {
		message = fmt.Sprintf("Task '%s' in project '%s' was due today and is now overdue", taskTitle, projectName)
	} else if daysPastDue == 1 {
		message = fmt.Sprintf("Task '%s' in project '%s' is 1 day overdue", taskTitle, projectName)
	} else {
		message = fmt.Sprintf("Task '%s' in project '%s' is %d days overdue", taskTitle, projectName, daysPastDue)
	}
	
	// Create JSON data with task information
	if taskID != uuid.Nil {
		jsonData := fmt.Sprintf(`{"task_id": "%s", "due_date": "%s", "days_overdue": %d}`, taskID.String(), dueDate.Format(time.RFC3339), daysPastDue)
		data = &jsonData
	}
	
	return CreateNotification(db, userID, "task_overdue", title, message, data)
}

// CreateTaskUnblockedNotification creates notification when a blocking task is completed
func CreateTaskUnblockedNotification(db *pgxpool.Pool, userID uuid.UUID, taskTitle string, projectName string, completedTaskTitle string, taskID uuid.UUID) error {
	var title, message string
	var data *string
	
	title = "Task Unblocked"
	message = fmt.Sprintf("Task '%s' in project '%s' can now be started - the blocking task '%s' has been completed", taskTitle, projectName, completedTaskTitle)
	
	// Create JSON data with task information
	if taskID != uuid.Nil {
		jsonData := fmt.Sprintf(`{"task_id": "%s", "completed_task_title": "%s"}`, taskID.String(), completedTaskTitle)
		data = &jsonData
	}
	
	return CreateNotification(db, userID, "task_unblocked", title, message, data)
}

// CreateWeeklyProjectSummaryNotification creates weekly summary notifications
func CreateWeeklyProjectSummaryNotification(db *pgxpool.Pool, userID uuid.UUID, upcomingTasks []Task) error {
	if len(upcomingTasks) == 0 {
		return nil
	}
	
	var title, message string
	var data *string
	
	title = "Weekly Project Summary"
	if len(upcomingTasks) == 1 {
		message = "You have 1 task with an upcoming deadline this week"
	} else {
		message = fmt.Sprintf("You have %d tasks with upcoming deadlines this week", len(upcomingTasks))
	}
	
	// Create JSON data with upcoming tasks
	taskIDs := make([]string, len(upcomingTasks))
	for i, task := range upcomingTasks {
		taskIDs[i] = task.ID.String()
	}
	jsonData := fmt.Sprintf(`{"task_ids": ["%s"], "task_count": %d}`, 
		fmt.Sprintf(`","`), len(upcomingTasks))
	data = &jsonData
	
	return CreateNotification(db, userID, "weekly_summary", title, message, data)
}

// NotifyTaskDeadlines checks for upcoming deadlines and creates notifications
func NotifyTaskDeadlines(db *pgxpool.Pool) error {
	// Get tasks due in the next 24 hours
	upcomingTasks, err := GetUpcomingTasksAllUsers(db, 24)
	if err != nil {
		return err
	}
	
	for _, task := range upcomingTasks {
		// Check if we already sent a notification for this task
		if !hasExistingDeadlineNotification(db, task.UserID, task.ID) {
			project, err := GetProjectByID(db, task.ProjectID, task.UserID)
			if err != nil {
				continue
			}
			
			CreateTaskDeadlineNotification(db, task.UserID, task.Title, project.Name, *task.DueDate, task.ID)
		}
	}
	
	return nil
}

// NotifyOverdueTasks checks for overdue tasks and creates notifications
func NotifyOverdueTasks(db *pgxpool.Pool) error {
	// Get all overdue tasks
	overdueTasks, err := GetOverdueTasksAllUsers(db)
	if err != nil {
		return err
	}
	
	for _, task := range overdueTasks {
		// Check if we already sent an overdue notification for this task
		if !hasExistingOverdueNotification(db, task.UserID, task.ID) {
			project, err := GetProjectByID(db, task.ProjectID, task.UserID)
			if err != nil {
				continue
			}
			
			CreateTaskOverdueNotification(db, task.UserID, task.Title, project.Name, *task.DueDate, task.ID)
		}
	}
	
	return nil
}

// Helper functions

func hasExistingDeadlineNotification(db *pgxpool.Pool, userID uuid.UUID, taskID uuid.UUID) bool {
	var count int
	err := db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM notifications 
		WHERE user_id = $1 AND type = 'task_deadline' AND data::jsonb->>'task_id' = $2 AND created_at > NOW() - INTERVAL '24 hours'`,
		userID, taskID.String()).Scan(&count)
	return err == nil && count > 0
}

func hasExistingOverdueNotification(db *pgxpool.Pool, userID uuid.UUID, taskID uuid.UUID) bool {
	var count int
	err := db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM notifications 
		WHERE user_id = $1 AND type = 'task_overdue' AND data::jsonb->>'task_id' = $2 AND created_at > NOW() - INTERVAL '24 hours'`,
		userID, taskID.String()).Scan(&count)
	return err == nil && count > 0
}

// GetUpcomingTasksAllUsers gets upcoming tasks for all users
func GetUpcomingTasksAllUsers(db *pgxpool.Pool, hours int) ([]Task, error) {
	futureTime := time.Now().Add(time.Hour * time.Duration(hours))
	
	query := `
		SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, due_date, completed, display_order, created_at, updated_at
		FROM tasks
		WHERE due_date BETWEEN $1 AND $2 AND completed = false
		ORDER BY due_date ASC
	`
	
	rows, err := db.Query(context.Background(), query, time.Now(), futureTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.Completed, &t.DisplayOrder, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// GetOverdueTasksAllUsers gets overdue tasks for all users
func GetOverdueTasksAllUsers(db *pgxpool.Pool) ([]Task, error) {
	query := `
		SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, due_date, completed, display_order, created_at, updated_at
		FROM tasks
		WHERE due_date < $1 AND completed = false
		ORDER BY due_date ASC
	`
	
	rows, err := db.Query(context.Background(), query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.Completed, &t.DisplayOrder, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}