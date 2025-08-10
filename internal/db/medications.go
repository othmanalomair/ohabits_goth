package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetMedicationsByUserID returns all active medications for a user
func GetMedicationsByUserID(db *pgxpool.Pool, userID uuid.UUID) ([]Medication, error) {
	query := `
		SELECT id, user_id, name, dosage, scheduled_days, times_per_day, 
		       time_intervals, duration_type, start_date, end_date, notes, 
		       is_active, created_at, updated_at
		FROM medications 
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medications []Medication
	for rows.Next() {
		var medication Medication

		err := rows.Scan(
			&medication.ID,
			&medication.UserID,
			&medication.Name,
			&medication.Dosage,
			&medication.ScheduledDays,
			&medication.TimesPerDay,
			&medication.TimeIntervals,
			&medication.DurationType,
			&medication.StartDate,
			&medication.EndDate,
			&medication.Notes,
			&medication.IsActive,
			&medication.CreatedAt,
			&medication.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		medications = append(medications, medication)
	}

	return medications, nil
}

// GetMedicationByID returns a medication by ID
func GetMedicationByID(db *pgxpool.Pool, medicationID, userID uuid.UUID) (*Medication, error) {
	query := `
		SELECT id, user_id, name, dosage, scheduled_days, times_per_day, 
		       time_intervals, duration_type, start_date, end_date, notes, 
		       is_active, created_at, updated_at
		FROM medications 
		WHERE id = $1 AND user_id = $2 AND is_active = true
	`

	var medication Medication

	err := db.QueryRow(context.Background(), query, medicationID, userID).Scan(
		&medication.ID,
		&medication.UserID,
		&medication.Name,
		&medication.Dosage,
		&medication.ScheduledDays,
		&medication.TimesPerDay,
		&medication.TimeIntervals,
		&medication.DurationType,
		&medication.StartDate,
		&medication.EndDate,
		&medication.Notes,
		&medication.IsActive,
		&medication.CreatedAt,
		&medication.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &medication, nil
}

// CreateMedication creates a new medication
func CreateMedication(db *pgxpool.Pool, medication *Medication) error {
	medication.ID = uuid.New()
	medication.IsActive = true
	medication.CreatedAt = time.Now()
	medication.UpdatedAt = time.Now()

	query := `
		INSERT INTO medications (id, user_id, name, dosage, scheduled_days, 
		                        times_per_day, time_intervals, duration_type, 
		                        start_date, end_date, notes, is_active, 
		                        created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := db.Exec(context.Background(), query,
		medication.ID,
		medication.UserID,
		medication.Name,
		medication.Dosage,
		medication.ScheduledDays,
		medication.TimesPerDay,
		medication.TimeIntervals,
		medication.DurationType,
		medication.StartDate,
		medication.EndDate,
		medication.Notes,
		medication.IsActive,
		medication.CreatedAt,
		medication.UpdatedAt,
	)

	return err
}

// UpdateMedication updates an existing medication
func UpdateMedication(db *pgxpool.Pool, medication *Medication) error {
	medication.UpdatedAt = time.Now()

	query := `
		UPDATE medications 
		SET name = $3, dosage = $4, scheduled_days = $5, times_per_day = $6, 
		    time_intervals = $7, duration_type = $8, start_date = $9, 
		    end_date = $10, notes = $11, updated_at = $12
		WHERE id = $1 AND user_id = $2
	`

	_, err := db.Exec(context.Background(), query,
		medication.ID,
		medication.UserID,
		medication.Name,
		medication.Dosage,
		medication.ScheduledDays,
		medication.TimesPerDay,
		medication.TimeIntervals,
		medication.DurationType,
		medication.StartDate,
		medication.EndDate,
		medication.Notes,
		medication.UpdatedAt,
	)

	return err
}

// DeleteMedication soft deletes a medication (sets is_active to false)
func DeleteMedication(db *pgxpool.Pool, medicationID, userID uuid.UUID) error {
	query := `
		UPDATE medications 
		SET is_active = false, updated_at = $3
		WHERE id = $1 AND user_id = $2
	`

	_, err := db.Exec(context.Background(), query, medicationID, userID, time.Now())
	return err
}

// LogMedication creates a medication log entry
func LogMedication(db *pgxpool.Pool, log *MedicationLog) error {
	log.ID = uuid.New()
	log.CreatedAt = time.Now()
	log.UpdatedAt = time.Now()

	query := `
		INSERT INTO medication_logs (id, medication_id, user_id, taken, 
		                           scheduled_time, actual_time, date, notes, 
		                           created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := db.Exec(context.Background(), query,
		log.ID,
		log.MedicationID,
		log.UserID,
		log.Taken,
		log.ScheduledTime,
		log.ActualTime,
		log.Date,
		log.Notes,
		log.CreatedAt,
		log.UpdatedAt,
	)

	return err
}

// GetMedicationLogsForDate returns medication logs for a specific date
func GetMedicationLogsForDate(db *pgxpool.Pool, userID uuid.UUID, date time.Time) ([]MedicationLog, error) {
	query := `
		SELECT id, medication_id, user_id, taken, scheduled_time, 
		       actual_time, date, notes, created_at, updated_at
		FROM medication_logs 
		WHERE user_id = $1 AND date = $2
		ORDER BY scheduled_time
	`

	rows, err := db.Query(context.Background(), query, userID, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []MedicationLog
	for rows.Next() {
		var log MedicationLog
		err := rows.Scan(
			&log.ID,
			&log.MedicationID,
			&log.UserID,
			&log.Taken,
			&log.ScheduledTime,
			&log.ActualTime,
			&log.Date,
			&log.Notes,
			&log.CreatedAt,
			&log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// CheckMedicationLogExists checks if a medication log exists for a specific medication, date, and dose
func CheckMedicationLogExists(db *pgxpool.Pool, medicationID, userID uuid.UUID, date time.Time, scheduledTime string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM medication_logs 
			WHERE medication_id = $1 AND user_id = $2 AND date = $3 AND scheduled_time = $4 AND taken = true
		)
	`

	var exists bool
	err := db.QueryRow(context.Background(), query, medicationID, userID, date.Format("2006-01-02"), scheduledTime).Scan(&exists)
	return exists, err
}

// DeleteMedicationLog removes a medication log entry
func DeleteMedicationLog(db *pgxpool.Pool, medicationID, userID uuid.UUID, date time.Time, scheduledTime string) error {
	query := `
		DELETE FROM medication_logs 
		WHERE medication_id = $1 AND user_id = $2 AND date = $3 AND scheduled_time = $4
	`

	_, err := db.Exec(context.Background(), query, medicationID, userID, date.Format("2006-01-02"), scheduledTime)
	return err
}