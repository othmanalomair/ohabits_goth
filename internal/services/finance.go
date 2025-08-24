package services

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"ohabits.com/internal/db"
)

// FinanceService handles automatic payment processing and financial tasks
type FinanceService struct {
	db        *pgxpool.Pool
	isRunning bool
}

// NewFinanceService creates a new finance service
func NewFinanceService(database *pgxpool.Pool) *FinanceService {
	return &FinanceService{
		db: database,
	}
}

// StartBackgroundProcessing starts the background financial processing
func (s *FinanceService) StartBackgroundProcessing() {
	if s.isRunning {
		log.Println("Finance service is already running")
		return
	}

	s.isRunning = true
	log.Println("Starting finance background service...")

	// Run immediately on start
	go s.processAutoPayments()

	// Set up ticker for daily processing at midnight
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	timeUntilMidnight := tomorrow.Sub(now)

	// First run at next midnight
	time.AfterFunc(timeUntilMidnight, func() {
		s.processAutoPayments()
		
		// Then run every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		go func() {
			for range ticker.C {
				if !s.isRunning {
					ticker.Stop()
					return
				}
				s.processAutoPayments()
			}
		}()
	})
}

// StopBackgroundProcessing stops the background financial processing
func (s *FinanceService) StopBackgroundProcessing() {
	s.isRunning = false
	log.Println("Stopping finance background service...")
}

// processAutoPayments processes all due auto-payments
func (s *FinanceService) processAutoPayments() {
	log.Println("Processing auto-payments...")
	
	// Get all active payments with auto-pay enabled
	payments, err := s.getAutoPayPaymentsDue()
	if err != nil {
		log.Printf("Error fetching auto-pay payments: %v", err)
		return
	}
	
	processed := 0
	for _, payment := range payments {
		if s.processPayment(payment) {
			processed++
		}
	}
	
	if processed > 0 {
		log.Printf("Processed %d auto-payments", processed)
	}
}

// getAutoPayPaymentsDue gets all recurring payments that are due today and have auto-pay enabled
func (s *FinanceService) getAutoPayPaymentsDue() ([]db.RecurringPayment, error) {
	today := time.Now()
	dayOfMonth := today.Day()
	
	query := `
		SELECT id, user_id, name, category, amount, frequency, due_date, start_date,
			   end_date, remaining_payments, total_amount, description, provider,
			   account_number, auto_pay, renewal_notice_days, price_history, metadata,
			   is_active, created_at, updated_at
		FROM recurring_payments 
		WHERE is_active = true 
		  AND auto_pay = true
		  AND due_date = $1
		  AND (end_date IS NULL OR end_date >= $2)
		  AND start_date <= $2
	`
	
	rows, err := s.db.Query(context.Background(), query, dayOfMonth, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var payments []db.RecurringPayment
	for rows.Next() {
		var payment db.RecurringPayment
		var priceHistory, metadata []byte
		
		err := rows.Scan(
			&payment.ID, &payment.UserID, &payment.Name, &payment.Category,
			&payment.Amount, &payment.Frequency, &payment.DueDate, &payment.StartDate,
			&payment.EndDate, &payment.RemainingPayments, &payment.TotalAmount,
			&payment.Description, &payment.Provider, &payment.AccountNumber,
			&payment.AutoPay, &payment.RenewalNoticeDays, &priceHistory, &metadata,
			&payment.IsActive, &payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning payment: %v", err)
			continue
		}
		
		payment.PriceHistory = priceHistory
		payment.Metadata = metadata
		
		payments = append(payments, payment)
	}
	
	return payments, nil
}

// processPayment processes a single auto-payment
func (s *FinanceService) processPayment(payment db.RecurringPayment) bool {
	today := time.Now()
	
	// Check if payment was already made today
	if s.wasPaymentMadeToday(payment.ID, payment.UserID, today) {
		return false // Already processed
	}
	
	// Create payment log
	paymentLog := db.PaymentLog{
		ID:                 uuid.New(),
		UserID:             payment.UserID,
		RecurringPaymentID: &payment.ID,
		Amount:             payment.Amount,
		PaymentDate:        today,
		DueDate:            today, // It's due today, so due date is today
		PaymentMethod:      stringPtr("auto_pay"),
		Notes:              stringPtr("Automatic payment processed"),
		IsLate:             false,
		LateFee:            0,
		CreatedAt:          today,
		PaymentName:        payment.Name,
	}
	
	// Save payment log to database
	err := s.createPaymentLog(paymentLog)
	if err != nil {
		log.Printf("Error creating auto-payment log for %s: %v", payment.Name, err)
		return false
	}
	
	// Update remaining payments if applicable
	if payment.RemainingPayments != nil && *payment.RemainingPayments > 0 {
		remaining := *payment.RemainingPayments - 1
		if remaining <= 0 {
			// Payment is completed, deactivate it
			s.deactivatePayment(payment.ID)
			log.Printf("Auto-payment %s completed and deactivated", payment.Name)
		} else {
			// Update remaining payments
			s.updateRemainingPayments(payment.ID, remaining)
		}
	}
	
	log.Printf("Auto-payment processed: %s (%.2f)", payment.Name, payment.Amount)
	return true
}

// wasPaymentMadeToday checks if a payment was already made today
func (s *FinanceService) wasPaymentMadeToday(paymentID, userID uuid.UUID, date time.Time) bool {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	query := `
		SELECT COUNT(*) 
		FROM payment_logs 
		WHERE recurring_payment_id = $1 
		  AND user_id = $2 
		  AND payment_date >= $3 
		  AND payment_date < $4
	`
	
	var count int
	err := s.db.QueryRow(context.Background(), query, paymentID, userID, startOfDay, endOfDay).Scan(&count)
	if err != nil {
		log.Printf("Error checking if payment was made today: %v", err)
		return false
	}
	
	return count > 0
}

// createPaymentLog creates a new payment log entry
func (s *FinanceService) createPaymentLog(log db.PaymentLog) error {
	query := `
		INSERT INTO payment_logs (id, user_id, recurring_payment_id, amount, payment_date, 
								  due_date, payment_method, notes, is_late, late_fee, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	
	_, err := s.db.Exec(context.Background(), query,
		log.ID, log.UserID, log.RecurringPaymentID, log.Amount, log.PaymentDate,
		log.DueDate, log.PaymentMethod, log.Notes, log.IsLate, log.LateFee, log.CreatedAt,
	)
	
	return err
}

// updateRemainingPayments updates the remaining payments for a recurring payment
func (s *FinanceService) updateRemainingPayments(paymentID uuid.UUID, remaining int) error {
	query := `UPDATE recurring_payments SET remaining_payments = $1, updated_at = $2 WHERE id = $3`
	_, err := s.db.Exec(context.Background(), query, remaining, time.Now(), paymentID)
	return err
}

// deactivatePayment deactivates a recurring payment
func (s *FinanceService) deactivatePayment(paymentID uuid.UUID) error {
	query := `UPDATE recurring_payments SET is_active = false, updated_at = $1 WHERE id = $2`
	_, err := s.db.Exec(context.Background(), query, time.Now(), paymentID)
	return err
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}