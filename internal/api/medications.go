package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"ohabits.com/internal/db"
)

// MedicationsHandler handles the medications page
func MedicationsHandler(w http.ResponseWriter, r *http.Request, dbPool *pgxpool.Pool) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	medications, err := db.GetMedicationsByUserID(dbPool, userID)
	if err != nil {
		http.Error(w, "Failed to fetch medications", http.StatusInternalServerError)
		return
	}

	data := struct {
		Medications []db.Medication
	}{
		Medications: medications,
	}

	// Return JSON response or render template based on request
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// AddMedicationHandler handles adding a new medication
func AddMedicationHandler(w http.ResponseWriter, r *http.Request, dbPool *pgxpool.Pool) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	medicationName := strings.TrimSpace(r.FormValue("medication_name"))
	dosage := strings.TrimSpace(r.FormValue("dosage"))
	durationType := r.FormValue("duration_type")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")
	notes := r.FormValue("notes")
	timesPerDayStr := r.FormValue("times_per_day")

	if medicationName == "" || dosage == "" || durationType == "" || startDateStr == "" || timesPerDayStr == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Parse times per day
	timesPerDay, err := strconv.Atoi(timesPerDayStr)
	if err != nil || timesPerDay < 1 || timesPerDay > 10 {
		http.Error(w, "Invalid times per day", http.StatusBadRequest)
		return
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	// Parse end date if provided
	var endDate *time.Time
	if endDateStr != "" && durationType == "limited" {
		parsedEndDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end date", http.StatusBadRequest)
			return
		}
		endDate = &parsedEndDate
	}

	// Parse scheduled days
	selectedDays := r.Form["days"]
	if len(selectedDays) == 0 {
		http.Error(w, "Please select at least one day", http.StatusBadRequest)
		return
	}

	// Convert day names to proper format
	dayMapping := map[string]string{
		"sunday":    "Sunday",
		"monday":    "Monday",
		"tuesday":   "Tuesday",
		"wednesday": "Wednesday",
		"thursday":  "Thursday",
		"friday":    "Friday",
		"saturday":  "Saturday",
	}

	var scheduledDays []string
	for _, day := range selectedDays {
		if properDay, exists := dayMapping[strings.ToLower(day)]; exists {
			scheduledDays = append(scheduledDays, properDay)
		}
	}

	scheduledDaysJSON, err := json.Marshal(scheduledDays)
	if err != nil {
		http.Error(w, "Error processing scheduled days", http.StatusInternalServerError)
		return
	}

	// Create default time intervals based on times per day
	timeIntervals := make([]string, timesPerDay)
	for i := 0; i < timesPerDay; i++ {
		timeIntervals[i] = "dose_" + strconv.Itoa(i+1)
	}

	medication := &db.Medication{
		UserID:        userID,
		Name:          medicationName,
		Dosage:        dosage,
		ScheduledDays: scheduledDaysJSON,
		TimesPerDay:   timesPerDay,
		TimeIntervals: timeIntervals,
		DurationType:  durationType,
		StartDate:     startDate,
		EndDate:       endDate,
		Notes:         func() *string { if notes == "" { return nil }; return &notes }(),
	}

	err = db.CreateMedication(dbPool, medication)
	if err != nil {
		http.Error(w, "Failed to create medication", http.StatusInternalServerError)
		return
	}

	// Return updated medications list
	MedicationsHandler(w, r, dbPool)
}

// EditMedicationHandler handles editing a medication
func EditMedicationHandler(w http.ResponseWriter, r *http.Request, dbPool *pgxpool.Pool) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)
	medicationIDStr := vars["id"]

	medicationID, err := uuid.Parse(medicationIDStr)
	if err != nil {
		http.Error(w, "Invalid medication ID", http.StatusBadRequest)
		return
	}

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	medicationName := strings.TrimSpace(r.FormValue("medication_name"))
	dosage := strings.TrimSpace(r.FormValue("dosage"))
	durationType := r.FormValue("duration_type")
	notes := r.FormValue("notes")
	timesPerDayStr := r.FormValue("times_per_day")

	if medicationName == "" || dosage == "" || durationType == "" || timesPerDayStr == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Parse times per day
	timesPerDay, err := strconv.Atoi(timesPerDayStr)
	if err != nil || timesPerDay < 1 || timesPerDay > 10 {
		http.Error(w, "Invalid times per day", http.StatusBadRequest)
		return
	}

	// Get existing medication
	medication, err := db.GetMedicationByID(dbPool, medicationID, userID)
	if err != nil {
		http.Error(w, "Medication not found", http.StatusNotFound)
		return
	}

	// Update medication fields
	medication.Name = medicationName
	medication.Dosage = dosage
	medication.DurationType = durationType
	medication.TimesPerDay = timesPerDay
	if notes == "" {
		medication.Notes = nil
	} else {
		medication.Notes = &notes
	}

	// Update time intervals if times per day changed
	if len(medication.TimeIntervals) != timesPerDay {
		timeIntervals := make([]string, timesPerDay)
		for i := 0; i < timesPerDay; i++ {
			timeIntervals[i] = "dose_" + strconv.Itoa(i+1)
		}
		medication.TimeIntervals = timeIntervals
	}

	err = db.UpdateMedication(dbPool, medication)
	if err != nil {
		http.Error(w, "Failed to update medication", http.StatusInternalServerError)
		return
	}

	// Return updated medications list
	MedicationsHandler(w, r, dbPool)
}

// DeleteMedicationHandler handles deleting a medication
func DeleteMedicationHandler(w http.ResponseWriter, r *http.Request, dbPool *pgxpool.Pool) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)
	medicationIDStr := vars["id"]

	medicationID, err := uuid.Parse(medicationIDStr)
	if err != nil {
		http.Error(w, "Invalid medication ID", http.StatusBadRequest)
		return
	}

	err = db.DeleteMedication(dbPool, medicationID, userID)
	if err != nil {
		http.Error(w, "Failed to delete medication", http.StatusInternalServerError)
		return
	}

	// Return updated medications list
	MedicationsHandler(w, r, dbPool)
}

// LogMedicationHandler handles logging medication intake
func LogMedicationHandler(w http.ResponseWriter, r *http.Request, dbPool *pgxpool.Pool) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)
	medicationIDStr := vars["id"]
	doseStr := r.URL.Query().Get("dose")

	medicationID, err := uuid.Parse(medicationIDStr)
	if err != nil {
		http.Error(w, "Invalid medication ID", http.StatusBadRequest)
		return
	}

	doseIndex, err := strconv.Atoi(doseStr)
	if err != nil || doseIndex < 0 {
		http.Error(w, "Invalid dose index", http.StatusBadRequest)
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	scheduledTime := "dose_" + strconv.Itoa(doseIndex+1)

	// Check if already logged
	exists, err := db.CheckMedicationLogExists(dbPool, medicationID, userID, today, scheduledTime)
	if err != nil {
		http.Error(w, "Error checking medication log", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	log := &db.MedicationLog{
		MedicationID:  medicationID,
		UserID:        userID,
		Taken:         !exists, // Toggle the taken status
		ScheduledTime: scheduledTime,
		ActualTime:    &now,
		Date:          today,
	}

	err = db.LogMedication(dbPool, log)
	if err != nil {
		http.Error(w, "Failed to log medication", http.StatusInternalServerError)
		return
	}

	// Return updated medications list
	MedicationsHandler(w, r, dbPool)
}