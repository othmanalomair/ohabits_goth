package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Project operations

func CreateProject(db *pgxpool.Pool, project Project, userID uuid.UUID) error {
	query := `
		INSERT INTO projects (user_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, userID, project.Name, project.Description, now, now)
	return err
}

func GetAllProjects(db *pgxpool.Pool, userID uuid.UUID) ([]Project, error) {
	query := `
		SELECT p.id, p.user_id, p.name, p.description, p.created_at, p.updated_at,
		       COUNT(CASE WHEN t.parent_task_id IS NULL THEN 1 END) as main_tasks,
		       COUNT(CASE WHEN t.parent_task_id IS NOT NULL THEN 1 END) as sub_tasks,
		       COUNT(CASE WHEN t.completed = true THEN 1 END) as completed_tasks
		FROM projects p
		LEFT JOIN tasks t ON p.id = t.project_id
		WHERE p.user_id = $1
		GROUP BY p.id, p.user_id, p.name, p.description, p.created_at, p.updated_at
		ORDER BY p.created_at DESC
	`
	
	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.MainTasks, &p.SubTasks, &p.CompletedTasks)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, rows.Err()
}

func GetProjectByID(db *pgxpool.Pool, projectID uuid.UUID, userID uuid.UUID) (*Project, error) {
	query := `
		SELECT p.id, p.user_id, p.name, p.description, p.created_at, p.updated_at,
		       COUNT(CASE WHEN t.parent_task_id IS NULL THEN 1 END) as main_tasks,
		       COUNT(CASE WHEN t.parent_task_id IS NOT NULL THEN 1 END) as sub_tasks,
		       COUNT(CASE WHEN t.completed = true THEN 1 END) as completed_tasks
		FROM projects p
		LEFT JOIN tasks t ON p.id = t.project_id
		WHERE p.id = $1 AND p.user_id = $2
		GROUP BY p.id, p.user_id, p.name, p.description, p.created_at, p.updated_at
	`
	
	var p Project
	err := db.QueryRow(context.Background(), query, projectID, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.MainTasks, &p.SubTasks, &p.CompletedTasks,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func UpdateProject(db *pgxpool.Pool, project Project, userID uuid.UUID) error {
	query := `
		UPDATE projects
		SET name = $3, description = $4, updated_at = $5
		WHERE id = $1 AND user_id = $2
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, project.ID, userID, project.Name, project.Description, now)
	return err
}

func DeleteProject(db *pgxpool.Pool, projectID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1 AND user_id = $2`
	_, err := db.Exec(context.Background(), query, projectID, userID)
	return err
}

// Task operations

func CreateTask(db *pgxpool.Pool, task Task, userID uuid.UUID) error {
	query := `
		INSERT INTO tasks (user_id, project_id, parent_task_id, title, description, status, priority, due_date, collapsed, display_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, userID, task.ProjectID, task.ParentTaskID, task.Title, task.Description, task.Status, task.Priority, task.DueDate, task.Collapsed, task.DisplayOrder, now, now)
	return err
}

func GetTasksByProject(db *pgxpool.Pool, projectID uuid.UUID, userID uuid.UUID) ([]Task, error) {
	query := `
		SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, due_date, completed, collapsed, display_order, created_at, updated_at
		FROM tasks
		WHERE project_id = $1 AND user_id = $2
		ORDER BY display_order ASC, id ASC
	`
	
	rows, err := db.Query(context.Background(), query, projectID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.Completed, &t.Collapsed, &t.DisplayOrder, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		
		// Check if task is blocked by incomplete dependencies
		t.IsBlocked = isTaskBlocked(db, t.ID)
		
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// ToggleTaskCollapsed toggles the collapsed state of a task
func ToggleTaskCollapsed(db *pgxpool.Pool, taskID uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE tasks
		SET collapsed = NOT collapsed, updated_at = $3
		WHERE id = $1 AND user_id = $2
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, taskID, userID, now)
	return err
}

// SetTaskCollapsed sets the collapsed state of a task to a specific value
func SetTaskCollapsed(db *pgxpool.Pool, taskID uuid.UUID, userID uuid.UUID, collapsed bool) error {
	query := `
		UPDATE tasks
		SET collapsed = $3, updated_at = $4
		WHERE id = $1 AND user_id = $2
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, taskID, userID, collapsed, now)
	return err
}

// UpdateTaskOrder updates the display order of tasks
func UpdateTaskOrder(db *pgxpool.Pool, taskID uuid.UUID, newOrder int, userID uuid.UUID) error {
	query := `
		UPDATE tasks
		SET display_order = $3, updated_at = $4
		WHERE id = $1 AND user_id = $2
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, taskID, userID, newOrder, now)
	return err
}

// GetChildTasks returns all direct children of a task
func GetChildTasks(db *pgxpool.Pool, parentTaskID uuid.UUID, userID uuid.UUID) ([]Task, error) {
	query := `
		SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, due_date, completed, collapsed, display_order, created_at, updated_at
		FROM tasks
		WHERE parent_task_id = $1 AND user_id = $2
		ORDER BY display_order ASC, created_at ASC
	`
	
	rows, err := db.Query(context.Background(), query, parentTaskID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.Completed, &t.Collapsed, &t.DisplayOrder, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// GetAllChildTasks recursively returns all descendants of a task
func GetAllChildTasks(db *pgxpool.Pool, parentTaskID uuid.UUID, userID uuid.UUID) ([]Task, error) {
	var allChildren []Task
	
	// Get direct children
	directChildren, err := GetChildTasks(db, parentTaskID, userID)
	if err != nil {
		return nil, err
	}
	
	allChildren = append(allChildren, directChildren...)
	
	// Recursively get children of each direct child
	for _, child := range directChildren {
		grandChildren, err := GetAllChildTasks(db, child.ID, userID)
		if err != nil {
			return nil, err
		}
		allChildren = append(allChildren, grandChildren...)
	}
	
	return allChildren, nil
}

// MarkTaskAndChildrenComplete marks a task and all its children as completed
func MarkTaskAndChildrenComplete(db *pgxpool.Pool, taskID uuid.UUID, userID uuid.UUID) error {
	// Start transaction
	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Mark the parent task as completed
	_, err = tx.Exec(context.Background(), 
		`UPDATE tasks SET completed = true, status = 'Completed', updated_at = $3 WHERE id = $1 AND user_id = $2`,
		taskID, userID, time.Now())
	if err != nil {
		return err
	}

	// Get all children recursively
	allChildren, err := GetAllChildTasks(db, taskID, userID)
	if err != nil {
		return err
	}

	// Mark all children as completed
	for _, child := range allChildren {
		_, err = tx.Exec(context.Background(),
			`UPDATE tasks SET completed = true, status = 'Completed', updated_at = $3 WHERE id = $1 AND user_id = $2`,
			child.ID, userID, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
}

// MarkTaskAndChildrenIncomplete marks a task and all its children as incomplete
func MarkTaskAndChildrenIncomplete(db *pgxpool.Pool, taskID uuid.UUID, userID uuid.UUID) error {
	// Start transaction
	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Mark the parent task as incomplete
	_, err = tx.Exec(context.Background(), 
		`UPDATE tasks SET completed = false, status = 'Not Started', updated_at = $3 WHERE id = $1 AND user_id = $2`,
		taskID, userID, time.Now())
	if err != nil {
		return err
	}

	// Get all children recursively
	allChildren, err := GetAllChildTasks(db, taskID, userID)
	if err != nil {
		return err
	}

	// Mark all children as incomplete
	for _, child := range allChildren {
		_, err = tx.Exec(context.Background(),
			`UPDATE tasks SET completed = false, status = 'Not Started', updated_at = $3 WHERE id = $1 AND user_id = $2`,
			child.ID, userID, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
}

// CheckAndMarkParentComplete checks if all children are completed and marks parent as completed if so
func CheckAndMarkParentComplete(db *pgxpool.Pool, taskID uuid.UUID, userID uuid.UUID) error {
	// Get the task to find its parent
	task, err := GetTaskByID(db, taskID, userID)
	if err != nil {
		return err
	}

	// If no parent, nothing to check
	if task.ParentTaskID == nil {
		return nil
	}

	parentID := *task.ParentTaskID

	// Get all children of the parent
	children, err := GetChildTasks(db, parentID, userID)
	if err != nil {
		return err
	}

	// Check if all children are completed
	allCompleted := true
	for _, child := range children {
		if !child.Completed {
			allCompleted = false
			break
		}
	}

	// If all children are completed, mark parent as completed and collapse it
	if allCompleted && len(children) > 0 {
		_, err = db.Exec(context.Background(),
			`UPDATE tasks SET completed = true, status = 'Completed', collapsed = true, updated_at = $3 WHERE id = $1 AND user_id = $2`,
			parentID, userID, time.Now())
		if err != nil {
			return err
		}

		// Recursively check the parent's parent
		return CheckAndMarkParentComplete(db, parentID, userID)
	}

	return nil
}

// CheckAndMarkParentIncomplete checks if parent should be marked incomplete when a child becomes incomplete
func CheckAndMarkParentIncomplete(db *pgxpool.Pool, taskID uuid.UUID, userID uuid.UUID) error {
	// Get the task to find its parent
	task, err := GetTaskByID(db, taskID, userID)
	if err != nil {
		return err
	}

	// If no parent, nothing to check
	if task.ParentTaskID == nil {
		return nil
	}

	parentID := *task.ParentTaskID

	// Get the parent task
	parentTask, err := GetTaskByID(db, parentID, userID)
	if err != nil {
		return err
	}

	// If parent is already incomplete, nothing to do
	if !parentTask.Completed {
		return nil
	}

	// Parent is completed, but we have an incomplete child, so mark parent as incomplete
	_, err = db.Exec(context.Background(),
		`UPDATE tasks SET completed = false, status = 'Not Started', updated_at = $3 WHERE id = $1 AND user_id = $2`,
		parentID, userID, time.Now())
	if err != nil {
		return err
	}

	// Recursively check the parent's parent
	return CheckAndMarkParentIncomplete(db, parentID, userID)
}

func GetTaskHierarchy(db *pgxpool.Pool, projectID uuid.UUID, userID uuid.UUID) ([]TaskHierarchy, error) {
	// Get all tasks for the project
	tasks, err := GetTasksByProject(db, projectID, userID)
	if err != nil {
		return nil, err
	}

	// Build hierarchy map
	taskMap := make(map[uuid.UUID]*TaskHierarchy)
	var rootTasks []TaskHierarchy

	// First pass: create all task hierarchy objects
	for _, task := range tasks {
		th := TaskHierarchy{
			Task:     task,
			Level:    0,
			Children: []TaskHierarchy{},
		}
		taskMap[task.ID] = &th
	}

	// Second pass: build parent-child relationships
	for _, task := range tasks {
		th := taskMap[task.ID]
		if task.ParentTaskID == nil {
			// Root task
			rootTasks = append(rootTasks, *th)
		} else {
			// Child task
			if parent, exists := taskMap[*task.ParentTaskID]; exists {
				th.Level = parent.Level + 1
				parent.Children = append(parent.Children, *th)
			}
		}
	}

	return rootTasks, nil
}

func GetTaskByID(db *pgxpool.Pool, taskID uuid.UUID, userID uuid.UUID) (*Task, error) {
	query := `
		SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, due_date, completed, collapsed, display_order, created_at, updated_at
		FROM tasks
		WHERE id = $1 AND user_id = $2
	`
	
	var t Task
	err := db.QueryRow(context.Background(), query, taskID, userID).Scan(
		&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.Completed, &t.Collapsed, &t.DisplayOrder, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load dependencies, comments, and attachments
	t.Dependencies, _ = GetTaskDependencies(db, taskID)
	t.Comments, _ = GetTaskComments(db, taskID)
	t.Attachments, _ = GetTaskAttachments(db, taskID)
	t.IsBlocked = isTaskBlocked(db, taskID)

	return &t, nil
}

func UpdateTask(db *pgxpool.Pool, task Task, userID uuid.UUID) error {
	query := `
		UPDATE tasks
		SET title = $3, description = $4, status = $5, priority = $6, due_date = $7, completed = $8, collapsed = $9, display_order = $10, updated_at = $11
		WHERE id = $1 AND user_id = $2
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, task.ID, userID, task.Title, task.Description, task.Status, task.Priority, task.DueDate, task.Completed, task.Collapsed, task.DisplayOrder, now)
	return err
}

func UpdateTaskStatus(db *pgxpool.Pool, taskID uuid.UUID, status string, userID uuid.UUID) error {
	query := `
		UPDATE tasks
		SET status = $3, updated_at = $4
		WHERE id = $1 AND user_id = $2
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, taskID, userID, status, now)
	return err
}

func ToggleTaskCompletion(db *pgxpool.Pool, taskID uuid.UUID, userID uuid.UUID) error {
	// Check if task is blocked
	if isTaskBlocked(db, taskID) {
		return fmt.Errorf("cannot complete task: it has incomplete dependencies")
	}

	// Get current task state
	task, err := GetTaskByID(db, taskID, userID)
	if err != nil {
		return err
	}

	// Check if there are children for collapse/expand logic
	children, err := GetChildTasks(db, taskID, userID)
	if err != nil {
		return err
	}

	if task.Completed {
		// Task is currently completed, toggle to not completed
		if len(children) > 0 {
			// Mark this task and all children as incomplete, and expand the task
			err = MarkTaskAndChildrenIncomplete(db, taskID, userID)
			if err != nil {
				return err
			}
			// Expand the task when marking as incomplete
			_, err = db.Exec(context.Background(), 
				`UPDATE tasks SET collapsed = false, updated_at = $3 WHERE id = $1 AND user_id = $2`,
				taskID, userID, time.Now())
			if err != nil {
				return err
			}
			// Check if marking this task incomplete should mark parent as incomplete
			return CheckAndMarkParentIncomplete(db, taskID, userID)
		} else {
			// No children, just mark this task as incomplete
			query := `
				UPDATE tasks
				SET completed = false, status = 'Not Started', updated_at = $3
				WHERE id = $1 AND user_id = $2
			`
			now := time.Now()
			_, err := db.Exec(context.Background(), query, taskID, userID, now)
			if err != nil {
				return err
			}
			// Check if marking this task incomplete should mark parent as incomplete
			return CheckAndMarkParentIncomplete(db, taskID, userID)
		}
	} else {
		// Task is not completed, mark as completed
		if len(children) > 0 {
			// Mark this task and all children as completed, and collapse the task
			err = MarkTaskAndChildrenComplete(db, taskID, userID)
			if err != nil {
				return err
			}
			// Collapse the task when marking as complete
			_, err = db.Exec(context.Background(), 
				`UPDATE tasks SET collapsed = true, updated_at = $3 WHERE id = $1 AND user_id = $2`,
				taskID, userID, time.Now())
			return err
		} else {
			// No children, just mark this task as completed
			query := `
				UPDATE tasks
				SET completed = true, status = 'Completed', updated_at = $3
				WHERE id = $1 AND user_id = $2
			`
			now := time.Now()
			_, err := db.Exec(context.Background(), query, taskID, userID, now)
			if err != nil {
				return err
			}
		}

		// Check if completing this task should mark parent as completed
		return CheckAndMarkParentComplete(db, taskID, userID)
	}
}

func DeleteTask(db *pgxpool.Pool, taskID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1 AND user_id = $2`
	_, err := db.Exec(context.Background(), query, taskID, userID)
	return err
}

func MoveTask(db *pgxpool.Pool, taskID uuid.UUID, newParentID *uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE tasks
		SET parent_task_id = $3, updated_at = $4
		WHERE id = $1 AND user_id = $2
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, taskID, userID, newParentID, now)
	return err
}

// Task Dependencies

func CreateTaskDependency(db *pgxpool.Pool, taskID uuid.UUID, dependsOnTaskID uuid.UUID) error {
	query := `
		INSERT INTO task_dependencies (task_id, depends_on_task_id, created_at)
		VALUES ($1, $2, $3)
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, taskID, dependsOnTaskID, now)
	return err
}

func GetTaskDependencies(db *pgxpool.Pool, taskID uuid.UUID) ([]TaskDependency, error) {
	query := `
		SELECT td.id, td.task_id, td.depends_on_task_id, td.created_at,
		       t.title, t.status, t.completed
		FROM task_dependencies td
		JOIN tasks t ON td.depends_on_task_id = t.id
		WHERE td.task_id = $1
	`
	
	rows, err := db.Query(context.Background(), query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dependencies []TaskDependency
	for rows.Next() {
		var td TaskDependency
		var depTask Task
		err := rows.Scan(&td.ID, &td.TaskID, &td.DependsOnTaskID, &td.CreatedAt, &depTask.Title, &depTask.Status, &depTask.Completed)
		if err != nil {
			return nil, err
		}
		td.DependsOnTask = &depTask
		dependencies = append(dependencies, td)
	}

	return dependencies, rows.Err()
}

func DeleteTaskDependency(db *pgxpool.Pool, dependencyID uuid.UUID) error {
	query := `DELETE FROM task_dependencies WHERE id = $1`
	_, err := db.Exec(context.Background(), query, dependencyID)
	return err
}

func isTaskBlocked(db *pgxpool.Pool, taskID uuid.UUID) bool {
	query := `
		SELECT COUNT(*)
		FROM task_dependencies td
		JOIN tasks t ON td.depends_on_task_id = t.id
		WHERE td.task_id = $1 AND t.completed = false
	`
	
	var count int
	err := db.QueryRow(context.Background(), query, taskID).Scan(&count)
	if err != nil {
		return false
	}
	
	return count > 0
}

// Task Comments

func CreateTaskComment(db *pgxpool.Pool, comment TaskComment, userID uuid.UUID) error {
	query := `
		INSERT INTO task_comments (task_id, user_id, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, comment.TaskID, userID, comment.Comment, now, now)
	return err
}

func GetTaskComments(db *pgxpool.Pool, taskID uuid.UUID) ([]TaskComment, error) {
	query := `
		SELECT id, task_id, user_id, comment, created_at, updated_at
		FROM task_comments
		WHERE task_id = $1
		ORDER BY created_at ASC
	`
	
	rows, err := db.Query(context.Background(), query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []TaskComment
	for rows.Next() {
		var c TaskComment
		err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Comment, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, rows.Err()
}

// Task Attachments

func CreateTaskAttachment(db *pgxpool.Pool, attachment TaskAttachment, userID uuid.UUID) error {
	query := `
		INSERT INTO task_attachments (task_id, user_id, filename, file_path, file_size, mime_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	_, err := db.Exec(context.Background(), query, attachment.TaskID, userID, attachment.Filename, attachment.FilePath, attachment.FileSize, attachment.MimeType, now)
	return err
}

func GetTaskAttachments(db *pgxpool.Pool, taskID uuid.UUID) ([]TaskAttachment, error) {
	query := `
		SELECT id, task_id, user_id, filename, file_path, file_size, mime_type, created_at
		FROM task_attachments
		WHERE task_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := db.Query(context.Background(), query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []TaskAttachment
	for rows.Next() {
		var a TaskAttachment
		err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Filename, &a.FilePath, &a.FileSize, &a.MimeType, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}

	return attachments, rows.Err()
}

func GetTaskAttachmentByID(db *pgxpool.Pool, attachmentID uuid.UUID, userID uuid.UUID) (*TaskAttachment, error) {
	query := `
		SELECT id, task_id, user_id, filename, file_path, file_size, mime_type, created_at
		FROM task_attachments
		WHERE id = $1 AND user_id = $2
	`
	
	var a TaskAttachment
	err := db.QueryRow(context.Background(), query, attachmentID, userID).Scan(
		&a.ID, &a.TaskID, &a.UserID, &a.Filename, &a.FilePath, &a.FileSize, &a.MimeType, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	return &a, nil
}

func DeleteTaskAttachment(db *pgxpool.Pool, attachmentID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM task_attachments WHERE id = $1 AND user_id = $2`
	_, err := db.Exec(context.Background(), query, attachmentID, userID)
	return err
}

// Utility functions for notifications and filtering

func GetOverdueTasks(db *pgxpool.Pool, userID uuid.UUID) ([]Task, error) {
	query := `
		SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, due_date, completed, collapsed, display_order, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND due_date < $2 AND completed = false
		ORDER BY due_date ASC
	`
	
	rows, err := db.Query(context.Background(), query, userID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.Completed, &t.Collapsed, &t.DisplayOrder, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

func GetUpcomingTasks(db *pgxpool.Pool, userID uuid.UUID, hours int) ([]Task, error) {
	futureTime := time.Now().Add(time.Hour * time.Duration(hours))
	
	query := `
		SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, due_date, completed, collapsed, display_order, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND due_date BETWEEN $2 AND $3 AND completed = false
		ORDER BY due_date ASC
	`
	
	rows, err := db.Query(context.Background(), query, userID, time.Now(), futureTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.Completed, &t.Collapsed, &t.DisplayOrder, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}