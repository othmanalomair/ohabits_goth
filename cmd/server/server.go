package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"ohabits.com/internal/handlers"
	"ohabits.com/internal/util"
)

func Server() *http.Server {
	r := mux.NewRouter()

	r.HandleFunc("/login", handlers.LoginPage)
	r.HandleFunc("/signup", handlers.SignupPage)

	// Protected Routes (Auth)
	protected := r.NewRoute().Subrouter()
	protected.Use(util.AuthMiddleware)

	// Frontend routes (HTML)
	protected.HandleFunc("/", handlers.IndexHandler).Methods("GET")
	protected.HandleFunc("/habits_completions", handlers.HabitsCompletedByDate).Methods("GET")
	protected.HandleFunc("/habit/{id}/toggle", handlers.ToggleHabitCompletion).Methods("POST")
	protected.HandleFunc("/habits_by_day", handlers.HabitsByDay).Methods("GET")
	protected.HandleFunc("/todos", handlers.Todos).Methods("GET")
	protected.HandleFunc("/todos", handlers.NewTodo).Methods("POST")
	protected.HandleFunc("/todos/{id}/toggle", handlers.ToggleTodoCompletion).Methods("POST")
	protected.HandleFunc("/todos/{id}/delete", handlers.DeleteTodo).Methods("DELETE")
	protected.HandleFunc("/notes", handlers.Notes).Methods("GET")
	protected.HandleFunc("/notes", handlers.SaveNote).Methods("POST")
	protected.HandleFunc("/mood_rating", handlers.Mood).Methods("GET")
	protected.HandleFunc("/mood_rating", handlers.SaveMood).Methods("POST")
	protected.HandleFunc("/workout", handlers.GetWorkoutExercises).Methods("GET")
	protected.HandleFunc("/workout_loging", handlers.WorkoutLoging).Methods("GET")
	protected.HandleFunc("/workout_loging", handlers.SaveWorkoutLog).Methods("POST")
	protected.HandleFunc("/calendar", handlers.CalendarHandler).Methods("GET")
	protected.HandleFunc("/habits", handlers.HabitsPage).Methods("GET")
	protected.HandleFunc("/habits", handlers.AddHabit).Methods("POST")
	protected.HandleFunc("/habits/{id}/edit-form", handlers.EditHabitForm).Methods("GET")
	protected.HandleFunc("/habits/{id}/edit", handlers.EditHabit).Methods("POST")
	protected.HandleFunc("/habits/{id}/delete", handlers.DeleteHabit).Methods("POST")
	protected.HandleFunc("/habits/{id}/toggle", handlers.ToggleHabitDay).Methods("POST")
	protected.HandleFunc("/habits/{id}/cancel", handlers.CancelHabitEdit).Methods("GET")
	protected.HandleFunc("/view", handlers.ViewHandler).Methods("GET")
	protected.HandleFunc("/mnotes", handlers.NotesHandler).Methods("GET")
	protected.HandleFunc("/profile", handlers.ProfileHandler).Methods("GET")
	protected.HandleFunc("/profile", handlers.UpdateProfileHandler).Methods("POST")
	protected.HandleFunc("/signout", handlers.SignOutHandler).Methods("GET")

	// Workout Plans
	protected.HandleFunc("/workout_plan", handlers.WorkoutPlanPage).Methods("GET")
	protected.HandleFunc("/workout_plan", handlers.CreateWorkoutPlan).Methods("POST")
	protected.HandleFunc("/workout_plan/{id}/toggle", handlers.ToggleWorkoutPlan).Methods("GET")
	protected.HandleFunc("/workout_plan/{id}/delete", handlers.DeleteWorkoutPlan).Methods("POST")
	protected.HandleFunc("/workout_plan/{id}/edit-form", handlers.EditWorkoutPlanForm).Methods("GET")
	protected.HandleFunc("/workout_plan/{id}/edit", handlers.EditWorkoutPlan).Methods("POST")
	protected.HandleFunc("/workout_plan/{id}/update-day", handlers.UpdateWorkoutPlanDay).Methods("POST")
	protected.HandleFunc("/workout_plan/{id}/exercises", handlers.AddWorkoutExercise).Methods("POST")
	protected.HandleFunc("/workout_plan/{id}/save", handlers.SaveWorkoutPlan).Methods("POST")
	protected.HandleFunc("/workout_plan/{id}/cancel", handlers.CancelWorkoutPlanEdit).Methods("GET")

	// Workout Plan Exercise Routes
	protected.HandleFunc("/workout_plan/{id}/exercises/{order}/delete", handlers.DeleteWorkoutExercise).Methods("POST")
	protected.HandleFunc("/workout_plan/{id}/exercises/{order}/edit-form", handlers.EditWorkoutExerciseForm).Methods("GET")
	protected.HandleFunc("/workout_plan/{id}/exercises/{order}/edit", handlers.EditWorkoutExercise).Methods("POST")
	protected.HandleFunc("/workout_plan/{id}/exercises/{order}/cancel", handlers.CancelWorkoutExerciseEdit).Methods("GET")

	// Serve static files (css, js, etc.)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	fmt.Println("Server is set up on port 8080...")
	return srv
}
