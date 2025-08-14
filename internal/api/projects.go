package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

// Project API handlers

func GetProjects(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Projects []db.Project `json:"projects"`
	}{
		Projects: projects,
	})
}

func GetProjectStats(w http.ResponseWriter, r *http.Request) {
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

	// Get all tasks for this project
	tasks, err := db.GetTasksByProject(db.DB, projectID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate stats
	mainTasks := 0
	subTasks := 0
	completedTasks := 0

	for _, task := range tasks {
		if task.ParentTaskID == nil {
			mainTasks++
		} else {
			subTasks++
		}
		if task.Completed {
			completedTasks++
		}
	}

	totalTasks := mainTasks + subTasks

	stats := struct {
		MainTasks      int `json:"main_tasks"`
		SubTasks       int `json:"sub_tasks"`
		CompletedTasks int `json:"completed_tasks"`
		TotalTasks     int `json:"total_tasks"`
	}{
		MainTasks:      mainTasks,
		SubTasks:       subTasks,
		CompletedTasks: completedTasks,
		TotalTasks:     totalTasks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var project db.Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.CreateProject(db.DB, project, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Project created successfully"})
}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var project db.Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateProject(db.DB, project, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Project updated successfully"})
}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Project deleted successfully"})
}

// Task API handlers

func GetProjectTasks(w http.ResponseWriter, r *http.Request) {
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

	// Check if hierarchy is requested
	hierarchy := r.URL.Query().Get("hierarchy") == "true"
	
	if hierarchy {
		tasks, err := db.GetTaskHierarchy(db.DB, projectID, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Tasks []db.TaskHierarchy `json:"tasks"`
		}{
			Tasks: tasks,
		})
	} else {
		tasks, err := db.GetTasksByProject(db.DB, projectID, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Tasks []db.Task `json:"tasks"`
		}{
			Tasks: tasks,
		})
	}
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set default values
	if task.Status == "" {
		task.Status = "Not Started"
	}
	if task.Priority == "" {
		task.Priority = "None"
	}

	err = db.CreateTask(db.DB, task, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Task created successfully"})
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := db.GetTaskByID(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateTask(db.DB, task, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Task updated successfully"})
}

func ToggleTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	err = db.ToggleTaskCompletion(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Task toggled successfully"})
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	err = db.DeleteTask(db.DB, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Task deleted successfully"})
}

func MoveTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	type MoveRequest struct {
		NewParentID *uuid.UUID `json:"new_parent_id"`
	}

	var moveReq MoveRequest
	err = json.NewDecoder(r.Body).Decode(&moveReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.MoveTask(db.DB, taskID, moveReq.NewParentID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Task moved successfully"})
}

// Task Dependency handlers

func CreateTaskDependency(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	type DependencyRequest struct {
		DependsOnTaskID uuid.UUID `json:"depends_on_task_id"`
	}

	var depReq DependencyRequest
	err = json.NewDecoder(r.Body).Decode(&depReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.CreateTaskDependency(db.DB, taskID, depReq.DependsOnTaskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Task dependency created successfully"})
}

func DeleteTaskDependency(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dependencyID, err := uuid.Parse(params["dependencyId"])
	if err != nil {
		http.Error(w, "Invalid dependency ID", http.StatusBadRequest)
		return
	}

	err = db.DeleteTaskDependency(db.DB, dependencyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Task dependency deleted successfully"})
}

// Task Comment handlers

func CreateTaskComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var comment db.TaskComment
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment.TaskID = taskID
	err = db.CreateTaskComment(db.DB, comment, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Comment added successfully"})
}

// Task Attachment handlers

func CreateTaskAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	taskID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create attachment record
	attachment := db.TaskAttachment{
		TaskID:   taskID,
		Filename: header.Filename,
		FilePath: "/uploads/tasks/" + taskID.String() + "/" + header.Filename, // This would be actual file path
		FileSize: header.Size,
		MimeType: header.Header.Get("Content-Type"),
	}

	err = db.CreateTaskAttachment(db.DB, attachment, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "File uploaded successfully"})
}

func DeleteTaskAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	attachmentID, err := uuid.Parse(params["attachmentId"])
	if err != nil {
		http.Error(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	err = db.DeleteTaskAttachment(db.DB, attachmentID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Attachment deleted successfully"})
}

// Utility endpoints

func GetOverdueTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tasks, err := db.GetOverdueTasks(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Tasks []db.Task `json:"tasks"`
	}{
		Tasks: tasks,
	})
}

func GetUpcomingTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hoursStr := r.URL.Query().Get("hours")
	hours := 24 // default to 24 hours
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil {
			hours = h
		}
	}

	tasks, err := db.GetUpcomingTasks(db.DB, userID, hours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Tasks []db.Task `json:"tasks"`
	}{
		Tasks: tasks,
	})
}