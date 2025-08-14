package handlers

import (
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

// ProjectsPage renders the main projects page
func ProjectsPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projects, err := db.GetAllProjects(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Projects []db.Project
		User     db.User
	}{
		Projects: projects,
		User:     getUserFromContext(r),
	}

	err = tmpl.ExecuteTemplate(w, "projects", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateProject handles project creation
func CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	project := db.Project{
		Name:        name,
		Description: &description,
	}

	err := db.CreateProject(db.DB, project, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated projects list
	projects, err := db.GetAllProjects(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/partials/projects_list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, projects)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// EditProjectForm returns the edit form for a project
func EditProjectForm(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	projectID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := db.GetProjectByID(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/partials/project_edit_form.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// EditProject handles project updates
func EditProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	projectID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	project := db.Project{
		ID:          projectID,
		Name:        name,
		Description: &description,
	}

	err = db.UpdateProject(db.DB, project, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the updated project with stats
	updatedProject, err := db.GetProjectByID(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/partials/single_project.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, updatedProject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteProject handles project deletion
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	projectID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	err = db.DeleteProject(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated projects list
	projects, err := db.GetAllProjects(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/partials/projects_list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, projects)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ProjectDetails shows detailed view of a project with tasks
func ProjectDetails(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	params := mux.Vars(r)
	projectID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	project, err := db.GetProjectByID(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all tasks and build hierarchy with proper levels
	allTasks, err := db.GetTasksByProject(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build task map for level calculation
	taskMap := make(map[uuid.UUID]db.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	// Order tasks hierarchically and calculate proper levels
	orderedTasks := orderTasksHierarchically(allTasks)
	
	var taskHierarchies []db.TaskHierarchy
	for _, task := range orderedTasks {
		level := calculateTaskLevel(task, taskMap)
		
		// Check if this task has children
		hasChildren := false
		for _, t := range allTasks {
			if t.ParentTaskID != nil && *t.ParentTaskID == task.ID {
				hasChildren = true
				break
			}
		}
		
		// Check if this task should be hidden because its parent is collapsed
		shouldShow := true
		if task.ParentTaskID != nil {
			if parentTask, exists := taskMap[*task.ParentTaskID]; exists && parentTask.Collapsed {
				shouldShow = false
			}
		}
		
		if shouldShow {
			th := db.TaskHierarchy{
				Task:        task,
				Level:       level,
				HasChildren: hasChildren,
				Children:    []db.TaskHierarchy{},
			}
			taskHierarchies = append(taskHierarchies, th)
		}
	}

	data := struct {
		Project   *db.Project
		Tasks     []db.TaskHierarchy
		Filter    string
		User      db.User
	}{
		Project: project,
		Tasks:   taskHierarchies,
		Filter:  "",
		User:    getUserFromContext(r),
	}

	err = tmpl.ExecuteTemplate(w, "project_details", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateTask handles task creation
func CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	projectID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	status := r.FormValue("status")
	priority := r.FormValue("priority")
	dueDateStr := r.FormValue("due_date")
	parentTaskIDStr := r.FormValue("parent_task_id")

	if title == "" {
		http.Error(w, "Task title is required", http.StatusBadRequest)
		return
	}

	// Set default values if not provided
	if status == "" {
		status = "Not Started"
	}
	if priority == "" {
		priority = "None"
	}

	// Only set description if it's not empty
	var taskDescription *string
	if description != "" {
		taskDescription = &description
	}

	// Get next display order relative to siblings
	tasks, err := db.GetTasksByProject(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Count siblings (tasks with the same parent) to get relative display order
	var parentTaskID *uuid.UUID
	if parentTaskIDStr != "" {
		if parsed, err := uuid.Parse(parentTaskIDStr); err == nil {
			parentTaskID = &parsed
		}
	}
	
	maxDisplayOrder := 0
	for _, t := range tasks {
		// Check if they have the same parent (both nil or same UUID)
		sameParent := (parentTaskID == nil && t.ParentTaskID == nil) ||
			(parentTaskID != nil && t.ParentTaskID != nil && *parentTaskID == *t.ParentTaskID)
		
		if sameParent && t.DisplayOrder > maxDisplayOrder {
			maxDisplayOrder = t.DisplayOrder
		}
	}
	displayOrder := maxDisplayOrder + 1

	task := db.Task{
		ProjectID:    projectID,
		Title:        title,
		Description:  taskDescription,
		Status:       status,
		Priority:     priority,
		DisplayOrder: displayOrder,
	}

	// Parse due date if provided
	if dueDateStr != "" {
		dueDate, err := time.Parse("2006-01-02", dueDateStr)
		if err == nil {
			task.DueDate = &dueDate
		}
	}

	// Parse parent task ID if provided
	if parentTaskIDStr != "" {
		parentTaskID, err := uuid.Parse(parentTaskIDStr)
		if err == nil {
			task.ParentTaskID = &parentTaskID
		}
	}

	err = db.CreateTask(db.DB, task, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated task list with proper levels
	allTasks, err := db.GetTasksByProject(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build task map for level calculation
	taskMap := make(map[uuid.UUID]db.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	// Order tasks hierarchically and calculate proper levels, filtering collapsed children
	orderedTasks := orderTasksHierarchically(allTasks)
	var taskHierarchies []db.TaskHierarchy
	for _, task := range orderedTasks {
		level := calculateTaskLevel(task, taskMap)
		
		// Check if this task should be hidden because its parent is collapsed
		shouldShow := true
		if task.ParentTaskID != nil {
			if parentTask, exists := taskMap[*task.ParentTaskID]; exists && parentTask.Collapsed {
				shouldShow = false
			}
		}
		
		if shouldShow {
			// Check if this task has children
			hasChildren := false
			for _, t := range allTasks {
				if t.ParentTaskID != nil && *t.ParentTaskID == task.ID {
					hasChildren = true
					break
				}
			}
			
			th := db.TaskHierarchy{
				Task:        task,
				Level:       level,
				HasChildren: hasChildren,
				Children:    []db.TaskHierarchy{},
			}
			taskHierarchies = append(taskHierarchies, th)
		}
	}

	tmpl, err := template.ParseFiles("templates/partials/task_hierarchy.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, taskHierarchies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ToggleTaskCompletion handles task completion toggling
func ToggleTaskCompletion(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["taskId"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Get project ID before toggling
	task, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	projectID := task.ProjectID

	err = db.ToggleTaskCompletion(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Refresh the entire task list to show all affected changes (parent + children)
	refreshTaskList(w, projectID, userID)
}

// EditTaskForm returns the edit form for a task
func EditTaskForm(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["taskId"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all tasks in the same project for parent selection
	allTasks, err := db.GetTasksByProject(db.DB, task.ProjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Task     *db.Task
		AllTasks []db.Task
	}{
		Task:     task,
		AllTasks: allTasks,
	}

	tmpl, err := template.ParseFiles("templates/partials/task_edit_form.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// EditTask handles task updates
func EditTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["taskId"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	status := r.FormValue("status")
	priority := r.FormValue("priority")
	dueDateStr := r.FormValue("due_date")
	parentTaskIDStr := r.FormValue("parent_task_id")

	if title == "" {
		http.Error(w, "Task title is required", http.StatusBadRequest)
		return
	}

	// Get existing task to preserve project ID
	existingTask, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Only set description if it's not empty
	var taskDescription *string
	if description != "" {
		taskDescription = &description
	}

	task := db.Task{
		ID:          taskID,
		ProjectID:   existingTask.ProjectID,
		Title:       title,
		Description: taskDescription,
		Status:      status,
		Priority:    priority,
		Completed:   existingTask.Completed,
	}

	// Parse due date if provided
	if dueDateStr != "" {
		dueDate, err := time.Parse("2006-01-02", dueDateStr)
		if err == nil {
			task.DueDate = &dueDate
		}
	}

	// Parse parent task ID if provided
	if parentTaskIDStr != "" {
		parentTaskID, err := uuid.Parse(parentTaskIDStr)
		if err == nil {
			task.ParentTaskID = &parentTaskID
		}
	}

	err = db.UpdateTask(db.DB, task, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated task to return
	updatedTask, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render just the updated task with proper hierarchy context
	renderSingleTask(w, updatedTask.ID, userID)
}

// DeleteTask handles task deletion
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["taskId"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Get project ID before deleting task
	task, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projectID := task.ProjectID

	err = db.DeleteTask(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated task list with proper levels
	allTasks, err := db.GetTasksByProject(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build task map for level calculation
	taskMap := make(map[uuid.UUID]db.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	// Order tasks hierarchically and calculate proper levels, filtering collapsed children
	orderedTasks := orderTasksHierarchically(allTasks)
	var taskHierarchies []db.TaskHierarchy
	for _, task := range orderedTasks {
		level := calculateTaskLevel(task, taskMap)
		
		// Check if this task should be hidden because its parent is collapsed
		shouldShow := true
		if task.ParentTaskID != nil {
			if parentTask, exists := taskMap[*task.ParentTaskID]; exists && parentTask.Collapsed {
				shouldShow = false
			}
		}
		
		if shouldShow {
			// Check if this task has children
			hasChildren := false
			for _, t := range allTasks {
				if t.ParentTaskID != nil && *t.ParentTaskID == task.ID {
					hasChildren = true
					break
				}
			}
			
			th := db.TaskHierarchy{
				Task:        task,
				Level:       level,
				HasChildren: hasChildren,
				Children:    []db.TaskHierarchy{},
			}
			taskHierarchies = append(taskHierarchies, th)
		}
	}

	tmpl, err := template.ParseFiles("templates/partials/task_hierarchy.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, taskHierarchies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CancelProjectEdit returns the original project view
func CancelProjectEdit(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	projectID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := db.GetProjectByID(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/partials/single_project.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetSubTaskForm returns a form for creating a sub-task
func GetSubTaskForm(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	projectID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	parentTaskID := r.URL.Query().Get("parent")
	if parentTaskID == "" {
		http.Error(w, "Parent task ID is required", http.StatusBadRequest)
		return
	}

	parentUUID, err := uuid.Parse(parentTaskID)
	if err != nil {
		http.Error(w, "Invalid parent task ID", http.StatusBadRequest)
		return
	}

	// Verify parent task belongs to user and project
	parentTask, err := db.GetTaskByID(db.DB, parentUUID, userID)
	if err != nil {
		http.Error(w, "Parent task not found", http.StatusNotFound)
		return
	}

	if parentTask.ProjectID != projectID {
		http.Error(w, "Parent task does not belong to this project", http.StatusBadRequest)
		return
	}

	// Return sub-task creation form
	tmpl := `<div class="subtask-form">
		<form
			hx-post="/projects/` + projectID.String() + `/tasks"
			hx-target="#task-list"
			hx-swap="innerHTML"
			class="inline-subtask-form"
		>
			<input type="hidden" name="parent_task_id" value="` + parentTaskID + `" />
			<input
				type="text"
				name="title"
				placeholder="Add sub-task..."
				required
				autofocus
			/>
			<button type="submit">
				<img src="/static/images/svg/check-green.svg" alt="Add" />
			</button>
			<button
				type="button"
				onclick="this.closest('.subtask-form-container').innerHTML = ''"
			>
				<img src="/static/images/svg/x-red.svg" alt="Cancel" />
			</button>
		</form>
	</div>`

	t, err := template.New("subtask-form").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CancelTaskEdit returns the original task view
func CancelTaskEdit(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["taskId"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render just the canceled task with proper hierarchy context
	renderSingleTask(w, task.ID, userID)
}

// AddTaskComment handles adding comments to tasks
func AddTaskComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["taskId"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	comment := r.FormValue("comment")
	if comment == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	taskComment := db.TaskComment{
		TaskID:  taskID,
		Comment: comment,
	}

	err = db.CreateTaskComment(db.DB, taskComment, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated comments
	comments, err := db.GetTaskComments(db.DB, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/partials/task_comments.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, comments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Utility functions

func getUserFromContext(r *http.Request) db.User {
	userID, _ := r.Context().Value("userID").(uuid.UUID)
	user, _ := db.GetUserByID(db.DB, userID)
	if user != nil {
		return *user
	}
	return db.User{}
}

func filterTasks(tasks []db.TaskHierarchy, filterFunc func(db.Task) bool) []db.TaskHierarchy {
	var filtered []db.TaskHierarchy
	
	for _, task := range tasks {
		if filterFunc(task.Task) {
			// Include this task and filter its children
			newTask := task
			newTask.Children = filterTasks(task.Children, filterFunc)
			filtered = append(filtered, newTask)
		} else {
			// Check children and include them if they match
			childrenFiltered := filterTasks(task.Children, filterFunc)
			if len(childrenFiltered) > 0 {
				newTask := task
				newTask.Children = childrenFiltered
				filtered = append(filtered, newTask)
			}
		}
	}
	
	return filtered
}

func flattenHierarchy(hierarchy []db.TaskHierarchy) []db.TaskHierarchy {
	var flattened []db.TaskHierarchy
	
	var flatten func(tasks []db.TaskHierarchy)
	flatten = func(tasks []db.TaskHierarchy) {
		for _, task := range tasks {
			// Add the current task to the flattened list
			flattened = append(flattened, task)
			// Recursively add children
			if len(task.Children) > 0 {
				flatten(task.Children)
			}
		}
	}
	
	flatten(hierarchy)
	return flattened
}

func calculateTaskLevel(task db.Task, taskMap map[uuid.UUID]db.Task) int {
	level := 0
	currentTask := task
	
	// Traverse up the parent chain to calculate level
	for currentTask.ParentTaskID != nil {
		level++
		parentID := *currentTask.ParentTaskID
		if parent, exists := taskMap[parentID]; exists {
			currentTask = parent
		} else {
			break // Parent not found, stop traversing
		}
		
		// Safety check to prevent infinite loops
		if level > 10 {
			break
		}
	}
	
	return level
}

// orderTasksHierarchically orders tasks so that parents appear before their children
// while preserving the display_order within each level
func orderTasksHierarchically(tasks []db.Task) []db.Task {
	taskMap := make(map[uuid.UUID]db.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	
	var ordered []db.Task
	visited := make(map[uuid.UUID]bool)
	
	// Get root tasks and sort by display order
	var rootTasks []db.Task
	for _, task := range tasks {
		if task.ParentTaskID == nil {
			rootTasks = append(rootTasks, task)
		}
	}
	
	// Sort root tasks by display order to preserve user's intended order
	sort.Slice(rootTasks, func(i, j int) bool {
		return rootTasks[i].DisplayOrder < rootTasks[j].DisplayOrder
	})
	
	// Traverse depth-first, preserving order within each level
	for _, task := range rootTasks {
		if !visited[task.ID] {
			traverseTaskOrderPreserving(task, taskMap, visited, &ordered)
		}
	}
	
	// Add any remaining tasks that might have been orphaned
	for _, task := range tasks {
		if !visited[task.ID] {
			traverseTaskOrderPreserving(task, taskMap, visited, &ordered)
		}
	}
	
	return ordered
}

// traverseTask recursively adds a task and its children to the ordered list
func traverseTask(task db.Task, taskMap map[uuid.UUID]db.Task, visited map[uuid.UUID]bool, ordered *[]db.Task) {
	if visited[task.ID] {
		return
	}
	
	visited[task.ID] = true
	*ordered = append(*ordered, task)
	
	// Find and traverse children
	for _, childTask := range taskMap {
		if childTask.ParentTaskID != nil && *childTask.ParentTaskID == task.ID && !visited[childTask.ID] {
			traverseTask(childTask, taskMap, visited, ordered)
		}
	}
}

// traverseTaskOrderPreserving recursively adds a task and its children to the ordered list,
// preserving display order within each level
func traverseTaskOrderPreserving(task db.Task, taskMap map[uuid.UUID]db.Task, visited map[uuid.UUID]bool, ordered *[]db.Task) {
	if visited[task.ID] {
		return
	}
	
	visited[task.ID] = true
	*ordered = append(*ordered, task)
	
	// Find children and sort them by display order before traversing
	var children []db.Task
	for _, childTask := range taskMap {
		if childTask.ParentTaskID != nil && *childTask.ParentTaskID == task.ID && !visited[childTask.ID] {
			children = append(children, childTask)
		}
	}
	
	// Sort children by display order to preserve user's intended order
	sort.Slice(children, func(i, j int) bool {
		return children[i].DisplayOrder < children[j].DisplayOrder
	})
	
	// Traverse children in the correct order
	for _, child := range children {
		traverseTaskOrderPreserving(child, taskMap, visited, ordered)
	}
}

// MoveTaskUp handles moving a task up in the order
func MoveTaskUp(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["taskId"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Get the current task
	task, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all tasks in the same project and same level
	allTasks, err := db.GetTasksByProject(db.DB, task.ProjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter tasks at the same level and same parent
	var siblingTasks []db.Task
	for _, t := range allTasks {
		// Check if they have the same parent (both nil or same UUID)
		sameParent := (task.ParentTaskID == nil && t.ParentTaskID == nil) ||
			(task.ParentTaskID != nil && t.ParentTaskID != nil && *task.ParentTaskID == *t.ParentTaskID)
		
		if sameParent {
			siblingTasks = append(siblingTasks, t)
		}
	}

	// Sort siblings by display order
	sort.Slice(siblingTasks, func(i, j int) bool {
		return siblingTasks[i].DisplayOrder < siblingTasks[j].DisplayOrder
	})

	// Find current task position and swap with previous
	for i, t := range siblingTasks {
		if t.ID == taskID && i > 0 {
			// Swap display orders
			currentOrder := siblingTasks[i].DisplayOrder
			prevOrder := siblingTasks[i-1].DisplayOrder
			
			err = db.UpdateTaskOrder(db.DB, taskID, prevOrder, userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			err = db.UpdateTaskOrder(db.DB, siblingTasks[i-1].ID, currentOrder, userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
	}

	// Return updated task list
	refreshTaskList(w, task.ProjectID, userID)
}

// MoveTaskDown handles moving a task down in the order
func MoveTaskDown(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["taskId"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Get the current task
	task, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all tasks in the same project and same level
	allTasks, err := db.GetTasksByProject(db.DB, task.ProjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter tasks at the same level and same parent
	var siblingTasks []db.Task
	for _, t := range allTasks {
		// Check if they have the same parent (both nil or same UUID)
		sameParent := (task.ParentTaskID == nil && t.ParentTaskID == nil) ||
			(task.ParentTaskID != nil && t.ParentTaskID != nil && *task.ParentTaskID == *t.ParentTaskID)
		
		if sameParent {
			siblingTasks = append(siblingTasks, t)
		}
	}

	// Sort siblings by display order
	sort.Slice(siblingTasks, func(i, j int) bool {
		return siblingTasks[i].DisplayOrder < siblingTasks[j].DisplayOrder
	})

	// Find current task position and swap with next
	for i, t := range siblingTasks {
		if t.ID == taskID && i < len(siblingTasks)-1 {
			// Swap display orders
			currentOrder := siblingTasks[i].DisplayOrder
			nextOrder := siblingTasks[i+1].DisplayOrder
			
			err = db.UpdateTaskOrder(db.DB, taskID, nextOrder, userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			err = db.UpdateTaskOrder(db.DB, siblingTasks[i+1].ID, currentOrder, userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
	}

	// Return updated task list
	refreshTaskList(w, task.ProjectID, userID)
}

// Helper function to refresh the task list
func refreshTaskList(w http.ResponseWriter, projectID uuid.UUID, userID uuid.UUID) {
	allTasks, err := db.GetTasksByProject(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build task map for level calculation
	taskMap := make(map[uuid.UUID]db.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	// Order tasks hierarchically and calculate proper levels
	orderedTasks := orderTasksHierarchically(allTasks)
	var taskHierarchies []db.TaskHierarchy
	for _, task := range orderedTasks {
		level := calculateTaskLevel(task, taskMap)
		
		// Check if this task has children
		hasChildren := false
		for _, t := range allTasks {
			if t.ParentTaskID != nil && *t.ParentTaskID == task.ID {
				hasChildren = true
				break
			}
		}
		
		// Check if this task should be hidden because its parent is collapsed
		shouldShow := true
		if task.ParentTaskID != nil {
			if parentTask, exists := taskMap[*task.ParentTaskID]; exists && parentTask.Collapsed {
				shouldShow = false
			}
		}
		
		if shouldShow {
			th := db.TaskHierarchy{
				Task:        task,
				Level:       level,
				HasChildren: hasChildren,
				Children:    []db.TaskHierarchy{},
			}
			taskHierarchies = append(taskHierarchies, th)
		}
	}

	tmpl, err := template.ParseFiles("templates/partials/task_hierarchy.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, taskHierarchies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ToggleTaskCollapsed handles toggling task collapsed state
func ToggleTaskCollapsed(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["taskId"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	err = db.ToggleTaskCollapsed(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get project ID to refresh the task list
	task, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated task list with proper levels
	allTasks, err := db.GetTasksByProject(db.DB, task.ProjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build task map for level calculation
	taskMap := make(map[uuid.UUID]db.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	// Order tasks hierarchically and calculate proper levels, filtering collapsed children
	orderedTasks := orderTasksHierarchically(allTasks)
	var taskHierarchies []db.TaskHierarchy
	for _, task := range orderedTasks {
		level := calculateTaskLevel(task, taskMap)
		
		// Check if this task should be hidden because its parent is collapsed
		shouldShow := true
		if task.ParentTaskID != nil {
			if parentTask, exists := taskMap[*task.ParentTaskID]; exists && parentTask.Collapsed {
				shouldShow = false
			}
		}
		
		if shouldShow {
			// Check if this task has children
			hasChildren := false
			for _, childTask := range allTasks {
				if childTask.ParentTaskID != nil && *childTask.ParentTaskID == task.ID {
					hasChildren = true
					break
				}
			}
			
			th := db.TaskHierarchy{
				Task:        task,
				Level:       level,
				HasChildren: hasChildren,
				Children:    []db.TaskHierarchy{},
			}
			taskHierarchies = append(taskHierarchies, th)
		}
	}

	tmpl, err := template.ParseFiles("templates/partials/task_hierarchy.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, taskHierarchies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ReorderTasks handles drag-and-drop reordering of tasks
func ReorderTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	projectID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Parse the reordered task IDs from the request
	taskIDsStr := r.FormValue("task_ids")
	if taskIDsStr == "" {
		http.Error(w, "Task IDs required", http.StatusBadRequest)
		return
	}

	// Split the comma-separated task IDs
	taskIDStrings := strings.Split(taskIDsStr, ",")
	
	// Update the display order for each task
	for i, taskIDStr := range taskIDStrings {
		taskID, err := uuid.Parse(strings.TrimSpace(taskIDStr))
		if err != nil {
			continue // Skip invalid UUIDs
		}
		
		err = db.UpdateTaskOrder(db.DB, taskID, i+1, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Return updated task list
	allTasks, err := db.GetTasksByProject(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build task map for level calculation
	taskMap := make(map[uuid.UUID]db.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	// Order tasks hierarchically and calculate proper levels
	orderedTasks := orderTasksHierarchically(allTasks)
	var taskHierarchies []db.TaskHierarchy
	for _, task := range orderedTasks {
		level := calculateTaskLevel(task, taskMap)
		
		// Check if this task has children
		hasChildren := false
		for _, t := range allTasks {
			if t.ParentTaskID != nil && *t.ParentTaskID == task.ID {
				hasChildren = true
				break
			}
		}
		
		th := db.TaskHierarchy{
			Task:        task,
			Level:       level,
			HasChildren: hasChildren,
			Children:    []db.TaskHierarchy{},
		}
		taskHierarchies = append(taskHierarchies, th)
	}

	tmpl, err := template.ParseFiles("templates/partials/task_hierarchy.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, taskHierarchies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
// renderSingleTask renders a single task for HTMX partial updates
func renderSingleTask(w http.ResponseWriter, taskID uuid.UUID, userID uuid.UUID) {
	// Get the task
	task, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all tasks in the project to calculate hierarchy
	allTasks, err := db.GetTasksByProject(db.DB, task.ProjectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build task map for level calculation
	taskMap := make(map[uuid.UUID]db.Task)
	for _, t := range allTasks {
		taskMap[t.ID] = t
	}

	// Calculate level and check if it has children
	level := calculateTaskLevel(*task, taskMap)
	hasChildren := false
	for _, t := range allTasks {
		if t.ParentTaskID != nil && *t.ParentTaskID == task.ID {
			hasChildren = true
			break
		}
	}

	// Create TaskHierarchy for the single task
	taskHierarchy := db.TaskHierarchy{
		Task:        *task,
		Level:       level,
		HasChildren: hasChildren,
		Children:    []db.TaskHierarchy{},
	}

	// Use the existing task_hierarchy template to render just this single task
	tmpl, err := template.ParseFiles("templates/partials/task_hierarchy.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Pass a slice with just this single task
	err = tmpl.Execute(w, []db.TaskHierarchy{taskHierarchy})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
