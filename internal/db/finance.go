package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetUserFinance retrieves user's financial profile
func GetUserFinance(userID uuid.UUID) (*UserFinance, error) {
	var finance UserFinance
	query := `
		SELECT id, user_id, monthly_income, emergency_buffer_amount, 
			   emergency_buffer_percentage, currency, created_at, updated_at
		FROM user_finances
		WHERE user_id = $1`
	
	err := DB.QueryRow(context.Background(), query, userID).Scan(
		&finance.ID, &finance.UserID, &finance.MonthlyIncome,
		&finance.EmergencyBufferAmount, &finance.EmergencyBufferPercent,
		&finance.Currency, &finance.CreatedAt, &finance.UpdatedAt,
	)
	
	if err == pgx.ErrNoRows {
		// Create default finance profile if none exists
		return CreateUserFinance(userID)
	}
	
	if err != nil {
		return nil, fmt.Errorf("error getting user finance: %v", err)
	}
	
	return &finance, nil
}

// CreateUserFinance creates a new financial profile for user
func CreateUserFinance(userID uuid.UUID) (*UserFinance, error) {
	finance := UserFinance{
		ID:                     uuid.New(),
		UserID:                 userID,
		MonthlyIncome:         0,
		EmergencyBufferAmount: 0,
		EmergencyBufferPercent: 10.0,
		Currency:              "USD",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	
	query := `
		INSERT INTO user_finances (id, user_id, monthly_income, emergency_buffer_amount, 
								  emergency_buffer_percentage, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := DB.Exec(context.Background(), query, finance.ID, finance.UserID, finance.MonthlyIncome,
		finance.EmergencyBufferAmount, finance.EmergencyBufferPercent,
		finance.Currency, finance.CreatedAt, finance.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("error creating user finance: %v", err)
	}
	
	return &finance, nil
}

// UpdateUserFinance updates user's financial profile
func UpdateUserFinance(userID uuid.UUID, income float64, bufferAmount float64, bufferPercent float64, currency string) error {
	query := `
		UPDATE user_finances 
		SET monthly_income = $2, emergency_buffer_amount = $3, 
			emergency_buffer_percentage = $4, currency = $5, updated_at = NOW()
		WHERE user_id = $1`
	
	_, err := DB.Exec(context.Background(), query, userID, income, bufferAmount, bufferPercent, currency)
	if err != nil {
		return fmt.Errorf("error updating user finance: %v", err)
	}
	
	return nil
}

// GetRecurringPayments retrieves all active recurring payments for a user
func GetRecurringPayments(userID uuid.UUID) ([]RecurringPayment, error) {
	query := `
		SELECT id, user_id, name, category, amount, frequency, due_date, start_date,
			   end_date, remaining_payments, total_amount, description, provider,
			   account_number, auto_pay, renewal_notice_days, price_history, metadata,
			   is_active, created_at, updated_at
		FROM recurring_payments
		WHERE user_id = $1 AND is_active = true
		ORDER BY due_date ASC, name ASC`
	
	rows, err := DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting recurring payments: %v", err)
	}
	defer rows.Close()
	
	var payments []RecurringPayment
	for rows.Next() {
		var payment RecurringPayment
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
			return nil, fmt.Errorf("error scanning recurring payment: %v", err)
		}
		
		payment.PriceHistory = priceHistory
		payment.Metadata = metadata
		
		// Calculate next due date and days until due
		calculatePaymentDates(&payment)
		
		payments = append(payments, payment)
	}
	
	return payments, nil
}

// CreateRecurringPayment creates a new recurring payment
func CreateRecurringPayment(payment *RecurringPayment) error {
	payment.ID = uuid.New()
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()
	payment.IsActive = true
	
	if payment.PriceHistory == nil {
		payment.PriceHistory = []byte("[]")
	}
	if payment.Metadata == nil {
		payment.Metadata = []byte("{}")
	}
	
	query := `
		INSERT INTO recurring_payments (
			id, user_id, name, category, amount, frequency, due_date, start_date,
			end_date, remaining_payments, total_amount, description, provider,
			account_number, auto_pay, renewal_notice_days, price_history, metadata,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`
	
	_, err := DB.Exec(context.Background(), query,
		payment.ID, payment.UserID, payment.Name, payment.Category,
		payment.Amount, payment.Frequency, payment.DueDate, payment.StartDate,
		payment.EndDate, payment.RemainingPayments, payment.TotalAmount,
		payment.Description, payment.Provider, payment.AccountNumber,
		payment.AutoPay, payment.RenewalNoticeDays, payment.PriceHistory,
		payment.Metadata, payment.IsActive, payment.CreatedAt, payment.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("error creating recurring payment: %v", err)
	}
	
	return nil
}

// UpdateRecurringPayment updates an existing recurring payment
func UpdateRecurringPayment(payment *RecurringPayment) error {
	payment.UpdatedAt = time.Now()
	
	query := `
		UPDATE recurring_payments 
		SET name = $3, category = $4, amount = $5, frequency = $6, due_date = $7,
			end_date = $8, remaining_payments = $9, total_amount = $10, description = $11,
			provider = $12, account_number = $13, auto_pay = $14, renewal_notice_days = $15,
			price_history = $16, metadata = $17, updated_at = $18
		WHERE id = $1 AND user_id = $2`
	
	_, err := DB.Exec(context.Background(), query,
		payment.ID, payment.UserID, payment.Name, payment.Category,
		payment.Amount, payment.Frequency, payment.DueDate,
		payment.EndDate, payment.RemainingPayments, payment.TotalAmount,
		payment.Description, payment.Provider, payment.AccountNumber,
		payment.AutoPay, payment.RenewalNoticeDays, payment.PriceHistory,
		payment.Metadata, payment.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("error updating recurring payment: %v", err)
	}
	
	return nil
}

// DeleteRecurringPayment soft deletes a recurring payment
func DeleteRecurringPayment(paymentID, userID uuid.UUID) error {
	query := `UPDATE recurring_payments SET is_active = false WHERE id = $1 AND user_id = $2`
	_, err := DB.Exec(context.Background(), query, paymentID, userID)
	if err != nil {
		return fmt.Errorf("error deleting recurring payment: %v", err)
	}
	return nil
}

// GetSavingsGoals retrieves all active savings goals for a user
func GetSavingsGoals(userID uuid.UUID) ([]SavingsGoal, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, target_date, priority,
			   monthly_contribution, auto_calculate_contribution, product_url, image_url,
			   description, is_achieved, achieved_date, is_active, is_archived, created_at, updated_at
		FROM savings_goals
		WHERE user_id = $1 AND is_active = true AND is_archived = false
		ORDER BY priority DESC, target_date ASC NULLS LAST, name ASC`
	
	rows, err := DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting savings goals: %v", err)
	}
	defer rows.Close()
	
	var goals []SavingsGoal
	for rows.Next() {
		var goal SavingsGoal
		
		err := rows.Scan(
			&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount,
			&goal.CurrentAmount, &goal.TargetDate, &goal.Priority,
			&goal.MonthlyContribution, &goal.AutoCalculateContrib,
			&goal.ProductURL, &goal.ImageURL, &goal.Description,
			&goal.IsAchieved, &goal.AchievedDate, &goal.IsActive, &goal.IsArchived,
			&goal.CreatedAt, &goal.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning savings goal: %v", err)
		}
		
		// Calculate progress and requirements
		calculateGoalProgress(&goal)
		
		goals = append(goals, goal)
	}
	
	return goals, nil
}

// GetArchivedSavingsGoals retrieves all archived savings goals for a user
func GetArchivedSavingsGoals(userID uuid.UUID) ([]SavingsGoal, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, target_date, priority,
			   monthly_contribution, auto_calculate_contribution, product_url, image_url,
			   description, is_achieved, achieved_date, is_active, is_archived, created_at, updated_at
		FROM savings_goals
		WHERE user_id = $1 AND is_archived = true
		ORDER BY achieved_date DESC NULLS LAST, created_at DESC`
	
	rows, err := DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting archived savings goals: %v", err)
	}
	defer rows.Close()
	
	var goals []SavingsGoal
	for rows.Next() {
		var goal SavingsGoal
		
		err := rows.Scan(
			&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount,
			&goal.CurrentAmount, &goal.TargetDate, &goal.Priority,
			&goal.MonthlyContribution, &goal.AutoCalculateContrib,
			&goal.ProductURL, &goal.ImageURL, &goal.Description,
			&goal.IsAchieved, &goal.AchievedDate, &goal.IsActive, &goal.IsArchived,
			&goal.CreatedAt, &goal.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning archived savings goal: %v", err)
		}
		
		// Calculate progress and requirements
		calculateGoalProgress(&goal)
		
		goals = append(goals, goal)
	}
	
	return goals, nil
}

// ArchiveSavingsGoal marks a savings goal as archived
func ArchiveSavingsGoal(goalID, userID uuid.UUID) error {
	query := `
		UPDATE savings_goals 
		SET is_archived = true, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND is_achieved = true`
	
	result, err := DB.Exec(context.Background(), query, goalID, userID)
	if err != nil {
		return fmt.Errorf("error archiving savings goal: %v", err)
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("goal not found, not achieved, or already archived")
	}
	
	return nil
}

// UnarchiveSavingsGoal marks a savings goal as not archived
func UnarchiveSavingsGoal(goalID, userID uuid.UUID) error {
	query := `
		UPDATE savings_goals 
		SET is_archived = false, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`
	
	result, err := DB.Exec(context.Background(), query, goalID, userID)
	if err != nil {
		return fmt.Errorf("error unarchiving savings goal: %v", err)
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("goal not found")
	}
	
	return nil
}

// GetSavingsGoalByID gets a specific savings goal by ID for a user
func GetSavingsGoalByID(goalID, userID uuid.UUID) (*SavingsGoal, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, target_date, priority,
			   monthly_contribution, auto_calculate_contribution, product_url, image_url,
			   description, is_achieved, achieved_date, is_active, is_archived, created_at, updated_at
		FROM savings_goals
		WHERE id = $1 AND user_id = $2`
	
	var goal SavingsGoal
	err := DB.QueryRow(context.Background(), query, goalID, userID).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Name,
		&goal.TargetAmount,
		&goal.CurrentAmount,
		&goal.TargetDate,
		&goal.Priority,
		&goal.MonthlyContribution,
		&goal.AutoCalculateContrib,
		&goal.ProductURL,
		&goal.ImageURL,
		&goal.Description,
		&goal.IsAchieved,
		&goal.AchievedDate,
		&goal.IsActive,
		&goal.IsArchived,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("savings goal not found")
		}
		return nil, fmt.Errorf("error getting savings goal: %v", err)
	}
	
	calculateGoalProgress(&goal)
	return &goal, nil
}

// CreateSavingsGoal creates a new savings goal
func CreateSavingsGoal(goal *SavingsGoal) error {
	goal.ID = uuid.New()
	goal.CreatedAt = time.Now()
	goal.UpdatedAt = time.Now()
	goal.IsActive = true
	goal.IsAchieved = false
	goal.CurrentAmount = 0
	
	query := `
		INSERT INTO savings_goals (
			id, user_id, name, target_amount, current_amount, target_date, priority,
			monthly_contribution, auto_calculate_contribution, product_url, image_url,
			description, is_achieved, achieved_date, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
	
	_, err := DB.Exec(context.Background(), query,
		goal.ID, goal.UserID, goal.Name, goal.TargetAmount,
		goal.CurrentAmount, goal.TargetDate, goal.Priority,
		goal.MonthlyContribution, goal.AutoCalculateContrib,
		goal.ProductURL, goal.ImageURL, goal.Description,
		goal.IsAchieved, goal.AchievedDate, goal.IsActive,
		goal.CreatedAt, goal.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("error creating savings goal: %v", err)
	}
	
	return nil
}

// UpdateSavingsGoal updates an existing savings goal
func UpdateSavingsGoal(goal *SavingsGoal) error {
	goal.UpdatedAt = time.Now()
	
	query := `
		UPDATE savings_goals 
		SET name = $3, target_amount = $4, target_date = $5, priority = $6,
			monthly_contribution = $7, auto_calculate_contribution = $8, product_url = $9,
			image_url = $10, description = $11, updated_at = $12
		WHERE id = $1 AND user_id = $2`
	
	_, err := DB.Exec(context.Background(), query,
		goal.ID, goal.UserID, goal.Name, goal.TargetAmount,
		goal.TargetDate, goal.Priority, goal.MonthlyContribution,
		goal.AutoCalculateContrib, goal.ProductURL, goal.ImageURL,
		goal.Description, goal.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("error updating savings goal: %v", err)
	}
	
	return nil
}

// DeleteSavingsGoal soft deletes a savings goal by setting is_active to false
func DeleteSavingsGoal(goalID, userID uuid.UUID) error {
	query := `UPDATE savings_goals SET is_active = false, updated_at = NOW() WHERE id = $1 AND user_id = $2`
	
	result, err := DB.Exec(context.Background(), query, goalID, userID)
	if err != nil {
		return fmt.Errorf("error deleting savings goal: %v", err)
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("savings goal not found or already deleted")
	}
	
	return nil
}

// AddSavingsContribution adds money to a savings goal
func AddSavingsContribution(userID, goalID uuid.UUID, amount float64, method, notes string) error {
	return AddSavingsContributionWithDate(userID, goalID, amount, time.Now(), method, notes)
}

func AddSavingsContributionWithDate(userID, goalID uuid.UUID, amount float64, contributionDate time.Time, method, notes string) error {
	tx, err := DB.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback(context.Background())
	
	// Insert contribution record
	contributionID := uuid.New()
	contributionQuery := `
		INSERT INTO savings_contributions (id, user_id, savings_goal_id, amount, contribution_date, method, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`
	
	_, err = tx.Exec(context.Background(), contributionQuery, contributionID, userID, goalID, amount, contributionDate, method, notes)
	if err != nil {
		return fmt.Errorf("error inserting contribution: %v", err)
	}
	
	// Update goal's current amount
	updateQuery := `
		UPDATE savings_goals 
		SET current_amount = current_amount + $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`
	
	_, err = tx.Exec(context.Background(), updateQuery, goalID, userID, amount)
	if err != nil {
		return fmt.Errorf("error updating goal amount: %v", err)
	}
	
	// Check if goal is achieved
	var currentAmount, targetAmount float64
	checkQuery := `SELECT current_amount, target_amount FROM savings_goals WHERE id = $1`
	err = tx.QueryRow(context.Background(), checkQuery, goalID).Scan(&currentAmount, &targetAmount)
	if err != nil {
		return fmt.Errorf("error checking goal status: %v", err)
	}
	
	if currentAmount >= targetAmount {
		achievedQuery := `
			UPDATE savings_goals 
			SET is_achieved = true, achieved_date = NOW(), updated_at = NOW()
			WHERE id = $1`
		_, err = tx.Exec(context.Background(), achievedQuery, goalID)
		if err != nil {
			return fmt.Errorf("error marking goal as achieved: %v", err)
		}
	}
	
	return tx.Commit(context.Background())
}

// GetSavingsGoalContributions gets all contributions for a specific savings goal
func GetSavingsGoalContributions(goalID, userID uuid.UUID) ([]SavingsContribution, error) {
	query := `
		SELECT sc.id, sc.user_id, sc.savings_goal_id, sc.amount, sc.contribution_date, 
			   sc.method, sc.notes, sc.created_at
		FROM savings_contributions sc
		JOIN savings_goals sg ON sc.savings_goal_id = sg.id
		WHERE sc.savings_goal_id = $1 AND sc.user_id = $2 AND sg.user_id = $2
		ORDER BY sc.contribution_date DESC, sc.created_at DESC`
	
	rows, err := DB.Query(context.Background(), query, goalID, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying contributions: %v", err)
	}
	defer rows.Close()
	
	var contributions []SavingsContribution
	for rows.Next() {
		var contribution SavingsContribution
		err := rows.Scan(
			&contribution.ID,
			&contribution.UserID,
			&contribution.SavingsGoalID,
			&contribution.Amount,
			&contribution.ContributionDate,
			&contribution.Method,
			&contribution.Notes,
			&contribution.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning contribution: %v", err)
		}
		contributions = append(contributions, contribution)
	}
	
	return contributions, nil
}

// GetSavingsContributionByID gets a specific contribution by ID
func GetSavingsContributionByID(contributionID, userID uuid.UUID) (*SavingsContribution, error) {
	query := `
		SELECT sc.id, sc.user_id, sc.savings_goal_id, sc.amount, sc.contribution_date, 
			   sc.method, sc.notes, sc.created_at, sg.name as goal_name
		FROM savings_contributions sc
		JOIN savings_goals sg ON sc.savings_goal_id = sg.id
		WHERE sc.id = $1 AND sc.user_id = $2 AND sg.user_id = $2`
	
	var contribution SavingsContribution
	err := DB.QueryRow(context.Background(), query, contributionID, userID).Scan(
		&contribution.ID,
		&contribution.UserID,
		&contribution.SavingsGoalID,
		&contribution.Amount,
		&contribution.ContributionDate,
		&contribution.Method,
		&contribution.Notes,
		&contribution.CreatedAt,
		&contribution.GoalName,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("contribution not found")
		}
		return nil, fmt.Errorf("error getting contribution: %v", err)
	}
	
	return &contribution, nil
}

// UpdateSavingsContribution updates a contribution record and adjusts goal amount
func UpdateSavingsContribution(contributionID, userID uuid.UUID, amount float64, contributionDate time.Time, method, notes string) error {
	tx, err := DB.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback(context.Background())
	
	// Get old contribution to calculate difference
	var oldAmount float64
	var goalID uuid.UUID
	err = tx.QueryRow(context.Background(), 
		"SELECT amount, savings_goal_id FROM savings_contributions WHERE id = $1 AND user_id = $2",
		contributionID, userID).Scan(&oldAmount, &goalID)
	if err != nil {
		return fmt.Errorf("error getting old contribution: %v", err)
	}
	
	// Update contribution
	updateQuery := `
		UPDATE savings_contributions 
		SET amount = $3, contribution_date = $4, method = $5, notes = $6
		WHERE id = $1 AND user_id = $2`
	
	_, err = tx.Exec(context.Background(), updateQuery, contributionID, userID, amount, contributionDate, method, notes)
	if err != nil {
		return fmt.Errorf("error updating contribution: %v", err)
	}
	
	// Adjust goal's current amount
	amountDifference := amount - oldAmount
	adjustQuery := `
		UPDATE savings_goals 
		SET current_amount = current_amount + $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`
	
	_, err = tx.Exec(context.Background(), adjustQuery, goalID, userID, amountDifference)
	if err != nil {
		return fmt.Errorf("error adjusting goal amount: %v", err)
	}
	
	return tx.Commit(context.Background())
}

// DeleteSavingsContribution deletes a contribution and adjusts goal amount
func DeleteSavingsContribution(contributionID, userID uuid.UUID) error {
	tx, err := DB.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback(context.Background())
	
	// Get contribution amount and goal ID before deleting
	var amount float64
	var goalID uuid.UUID
	err = tx.QueryRow(context.Background(),
		"SELECT amount, savings_goal_id FROM savings_contributions WHERE id = $1 AND user_id = $2",
		contributionID, userID).Scan(&amount, &goalID)
	if err != nil {
		return fmt.Errorf("error getting contribution: %v", err)
	}
	
	// Delete contribution
	_, err = tx.Exec(context.Background(), 
		"DELETE FROM savings_contributions WHERE id = $1 AND user_id = $2",
		contributionID, userID)
	if err != nil {
		return fmt.Errorf("error deleting contribution: %v", err)
	}
	
	// Reduce goal's current amount
	adjustQuery := `
		UPDATE savings_goals 
		SET current_amount = current_amount - $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`
	
	_, err = tx.Exec(context.Background(), adjustQuery, goalID, userID, amount)
	if err != nil {
		return fmt.Errorf("error adjusting goal amount: %v", err)
	}
	
	return tx.Commit(context.Background())
}

// LogPayment records a payment made for a recurring payment
func LogPayment(userID uuid.UUID, recurringPaymentID *uuid.UUID, amount float64, paymentDate, dueDate time.Time, method, notes string) error {
	paymentLog := PaymentLog{
		ID:                 uuid.New(),
		UserID:             userID,
		RecurringPaymentID: recurringPaymentID,
		Amount:             amount,
		PaymentDate:        paymentDate,
		DueDate:            dueDate,
		PaymentMethod:      &method,
		Notes:              &notes,
		IsLate:             paymentDate.After(dueDate),
		LateFee:            0,
		CreatedAt:          time.Now(),
	}
	
	query := `
		INSERT INTO payment_logs (
			id, user_id, recurring_payment_id, amount, payment_date, due_date,
			payment_method, notes, is_late, late_fee, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	
	_, err := DB.Exec(context.Background(), query,
		paymentLog.ID, paymentLog.UserID, paymentLog.RecurringPaymentID,
		paymentLog.Amount, paymentLog.PaymentDate, paymentLog.DueDate,
		paymentLog.PaymentMethod, paymentLog.Notes, paymentLog.IsLate,
		paymentLog.LateFee, paymentLog.CreatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("error logging payment: %v", err)
	}
	
	return nil
}

// GetFinanceDashboard builds comprehensive dashboard data
func GetFinanceDashboard(userID uuid.UUID) (*FinanceDashboard, error) {
	// Get user finance profile
	userFinance, err := GetUserFinance(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user finance: %v", err)
	}
	
	// Get recurring payments
	payments, err := GetRecurringPayments(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting payments: %v", err)
	}
	
	// Get savings goals
	goals, err := GetSavingsGoals(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting goals: %v", err)
	}
	
	// Add currency to each savings goal for templates
	for i := range goals {
		goals[i].Currency = userFinance.Currency
	}
	
	// Get payment logs for current month to check payment status
	currentTime := time.Now()
	startOfMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
	
	// Get all payment logs for the user (not just this month)
	allPaymentLogs, err := GetPaymentLogs(userID)
	if err != nil {
		log.Printf("Error getting payment logs: %v", err)
		// Continue without payment status - don't fail the whole dashboard
	}
	
	// Create map of recurring payment ID -> payment covers this month
	paidThisMonth := make(map[uuid.UUID]bool)
	for _, paymentLog := range allPaymentLogs {
		if paymentLog.RecurringPaymentID != nil {
			// Check if this payment covers the current month based on due date
			// A payment covers this month if its due date falls within this month
			// OR if it was logged this month (for payments without clear due dates)
			dueInCurrentMonth := paymentLog.DueDate.Year() == currentTime.Year() && 
								  paymentLog.DueDate.Month() == currentTime.Month()
			loggedInCurrentMonth := paymentLog.PaymentDate.After(startOfMonth.Add(-time.Nanosecond)) && 
								    paymentLog.PaymentDate.Before(endOfMonth.Add(time.Nanosecond))
			
			if dueInCurrentMonth || loggedInCurrentMonth {
				paidThisMonth[*paymentLog.RecurringPaymentID] = true
			}
		}
	}
	
	// Create map of recurring payment ID -> payment is paid for specific due date
	paidForDueDate := make(map[uuid.UUID]map[string]bool)
	for _, paymentLog := range allPaymentLogs {
		if paymentLog.RecurringPaymentID != nil {
			if paidForDueDate[*paymentLog.RecurringPaymentID] == nil {
				paidForDueDate[*paymentLog.RecurringPaymentID] = make(map[string]bool)
			}
			// Mark this specific due date as paid
			dueDateKey := paymentLog.DueDate.Format("2006-01-02")
			paidForDueDate[*paymentLog.RecurringPaymentID][dueDateKey] = true
		}
	}

	// Add currency and payment status to each payment for templates
	for i := range payments {
		payments[i].Currency = userFinance.Currency
		payments[i].IsPaidThisMonth = paidThisMonth[payments[i].ID]
		
		// Check if the next due date is paid
		if payments[i].NextDueDate != nil {
			nextDueDateKey := payments[i].NextDueDate.Format("2006-01-02")
			if paidForDueDate[payments[i].ID] != nil {
				payments[i].IsPaidForNextDueDate = paidForDueDate[payments[i].ID][nextDueDateKey]
			}
		}
	}
	
	// Calculate dashboard metrics
	dashboard := &FinanceDashboard{
		UserFinance:        userFinance,
		ActiveSavingsGoals: goals,
	}
	
	// Calculate monthly commitments and categorize payments
	now := time.Now()
	currentMonth := now.Month()
	currentYear := now.Year()
	
	var monthlyCommitments float64
	var thisMonthPayments []RecurringPayment
	var upcomingPayments []RecurringPayment
	var nextPaymentDue *RecurringPayment
	
	for i, payment := range payments {
		// Calculate monthly equivalent
		monthlyAmount := calculateMonthlyAmount(payment.Amount, payment.Frequency)
		monthlyCommitments += monthlyAmount
		
		// Check if payment belongs to this month (due this month OR should have been paid this month)
		if isPaymentForThisMonth(payment, currentMonth, currentYear) {
			thisMonthPayments = append(thisMonthPayments, payment)
		}
		
		// Find next payment due
		if payment.NextDueDate != nil {
			if nextPaymentDue == nil || payment.NextDueDate.Before(*nextPaymentDue.NextDueDate) {
				nextPaymentDue = &payments[i]
			}
		}
		
		// Upcoming payments (next 30 days)
		if payment.NextDueDate != nil && payment.NextDueDate.Before(now.AddDate(0, 0, 30)) {
			upcomingPayments = append(upcomingPayments, payment)
		}
	}
	
	dashboard.MonthlyCommitments = monthlyCommitments
	dashboard.ThisMonthPayments = thisMonthPayments
	dashboard.UpcomingPayments = upcomingPayments
	dashboard.NextPaymentDue = nextPaymentDue
	
	// Calculate total savings contribution
	var totalSavingsContrib float64
	for _, goal := range goals {
		totalSavingsContrib += goal.MonthlyContribution
	}
	
	// Calculate safe to spend
	emergencyBuffer := userFinance.EmergencyBufferAmount
	if emergencyBuffer == 0 {
		emergencyBuffer = userFinance.MonthlyIncome * (userFinance.EmergencyBufferPercent / 100)
	}
	
	dashboard.SafeToSpend = userFinance.MonthlyIncome - monthlyCommitments - totalSavingsContrib - emergencyBuffer
	dashboard.TotalSavingsProgress = totalSavingsContrib
	dashboard.TotalSavingsContrib = totalSavingsContrib
	dashboard.CalculatedEmergencyBuffer = emergencyBuffer
	
	return dashboard, nil
}

// Helper functions

func calculatePaymentDates(payment *RecurringPayment) {
	now := time.Now()
	var nextDue time.Time
	
	switch payment.Frequency {
	case "monthly":
		// For monthly payments, calculate next actual due date
		nextDue = getNextMonthlyDate(payment.DueDate, now)
	case "yearly":
		nextDue = getNextYearlyDate(payment.DueDate, payment.StartDate, now)
	case "weekly":
		nextDue = getNextWeeklyDate(payment.DueDate, now) // DueDate is day of week
	case "quarterly":
		nextDue = getNextQuarterlyDate(payment.DueDate, payment.StartDate, now)
	}
	
	payment.NextDueDate = &nextDue
	
	// Calculate days until due
	days := int(nextDue.Sub(now).Hours() / 24)
	payment.DaysUntilDue = &days
	payment.IsOverdue = nextDue.Before(now)
	
	// For monthly payments, calculate current month overdue status
	if payment.Frequency == "monthly" && !payment.IsPaidThisMonth {
		currentYear, currentMonth, _ := now.Date()
		thisMonthDue := time.Date(currentYear, currentMonth, payment.DueDate, 0, 0, 0, 0, now.Location())
		
		// Check if this month's payment was supposed to be made and hasn't been
		if thisMonthDue.Before(now) && thisMonthDue.After(payment.StartDate.AddDate(0, 0, -1)) {
			payment.IsOverdueForCurrentMonth = true
			payment.CurrentMonthDueDate = &thisMonthDue
			payment.DaysOverdue = int(now.Sub(thisMonthDue).Hours() / 24)
		}
	}
}

func calculateGoalProgress(goal *SavingsGoal) {
	if goal.TargetAmount > 0 {
		goal.ProgressPercentage = (goal.CurrentAmount / goal.TargetAmount) * 100
		if goal.ProgressPercentage > 100 {
			goal.ProgressPercentage = 100
		}
	}
	
	goal.RemainingAmount = goal.TargetAmount - goal.CurrentAmount
	if goal.RemainingAmount < 0 {
		goal.RemainingAmount = 0
	}
	
	// Calculate time requirements
	if goal.TargetDate != nil && !goal.IsAchieved && goal.RemainingAmount > 0 {
		now := time.Now()
		daysUntilTarget := int(goal.TargetDate.Sub(now).Hours() / 24)
		
		if daysUntilTarget > 0 {
			months := daysUntilTarget / 30
			weeks := daysUntilTarget / 7
			
			goal.MonthsToTarget = &months
			goal.WeeksToTarget = &weeks
			
			goal.DailyRequired = goal.RemainingAmount / float64(daysUntilTarget)
			goal.WeeklyRequired = goal.RemainingAmount / float64(weeks)
			
			// Check if on track
			requiredMonthly := goal.RemainingAmount / float64(months)
			goal.IsOnTrack = goal.MonthlyContribution >= requiredMonthly
		}
	}
}

func calculateMonthlyAmount(amount float64, frequency string) float64 {
	switch frequency {
	case "monthly":
		return amount
	case "yearly":
		return amount / 12
	case "weekly":
		return amount * 4.33 // Average weeks per month
	case "quarterly":
		return amount / 3
	default:
		return amount
	}
}

func isPaymentDueThisMonth(payment RecurringPayment, month time.Month, year int) bool {
	if payment.NextDueDate == nil {
		return false
	}
	return payment.NextDueDate.Month() == month && payment.NextDueDate.Year() == year
}

func isPaymentForThisMonth(payment RecurringPayment, month time.Month, year int) bool {
	// For monthly payments, check if this payment should appear in this month's list
	if payment.Frequency != "monthly" {
		return false
	}
	
	now := time.Now()
	thisMonthDueDate := time.Date(year, month, payment.DueDate, 0, 0, 0, 0, now.Location())
	
	// Check if payment started before or on this month's due date
	if payment.StartDate.After(thisMonthDueDate) {
		return false
	}
	
	// Show if:
	// 1. Payment is due this month (future)
	// 2. Payment was due this month but is overdue 
	// 3. Payment was due this month and has been paid
	
	// Check if the due date falls within this month
	if thisMonthDueDate.Month() != month || thisMonthDueDate.Year() != year {
		return false
	}
	
	// Always show payments that belong to this month, regardless of payment status
	return true
}

func hasPaymentThisMonth(payment *RecurringPayment, from time.Time) bool {
	year, month, _ := from.Date()
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, from.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
	
	// Query to check if there's a payment log for this recurring payment in this month
	query := `
		SELECT COUNT(*) 
		FROM payment_logs 
		WHERE recurring_payment_id = $1 
		AND payment_date >= $2 
		AND payment_date <= $3`
	
	var count int
	err := DB.QueryRow(context.Background(), query, payment.ID, startOfMonth, endOfMonth).Scan(&count)
	if err != nil {
		// If there's an error, assume no payment was made
		return false
	}
	
	return count > 0
}

func getNextMonthlyDate(dayOfMonth int, from time.Time) time.Time {
	year, month, _ := from.Date()
	thisMonthDate := time.Date(year, month, dayOfMonth, 0, 0, 0, 0, from.Location())
	
	// If the due date is today or in the future this month, use this month
	if thisMonthDate.After(from) || thisMonthDate.Equal(from.Truncate(24*time.Hour)) {
		return thisMonthDate
	}
	
	// Otherwise, move to next month
	nextMonthDate := thisMonthDate.AddDate(0, 1, 0)
	
	// Handle months with fewer days
	if nextMonthDate.Day() != dayOfMonth {
		nextMonthDate = time.Date(nextMonthDate.Year(), nextMonthDate.Month()+1, 1, 0, 0, 0, 0, nextMonthDate.Location()).AddDate(0, 0, -1)
	}
	
	return nextMonthDate
}

func getNextYearlyDate(dayOfMonth int, startDate, from time.Time) time.Time {
	year := from.Year()
	month := startDate.Month()
	
	nextDate := time.Date(year, month, dayOfMonth, 0, 0, 0, 0, from.Location())
	
	if nextDate.Before(from) || nextDate.Equal(from) {
		nextDate = nextDate.AddDate(1, 0, 0)
	}
	
	return nextDate
}

func getNextWeeklyDate(dayOfWeek int, from time.Time) time.Time {
	daysUntil := (dayOfWeek - int(from.Weekday()) + 7) % 7
	if daysUntil == 0 {
		daysUntil = 7 // Next week
	}
	return from.AddDate(0, 0, daysUntil)
}

func getNextQuarterlyDate(dayOfMonth int, startDate, from time.Time) time.Time {
	// Find the next quarter month
	quarterMonths := []time.Month{
		startDate.Month(),
		startDate.Month() + 3,
		startDate.Month() + 6,
		startDate.Month() + 9,
	}
	
	for _, qMonth := range quarterMonths {
		if qMonth > 12 {
			qMonth -= 12
		}
		
		year := from.Year()
		if qMonth < from.Month() {
			year++
		}
		
		nextDate := time.Date(year, qMonth, dayOfMonth, 0, 0, 0, 0, from.Location())
		if nextDate.After(from) {
			return nextDate
		}
	}
	
	// Fallback to next year
	return time.Date(from.Year()+1, quarterMonths[0], dayOfMonth, 0, 0, 0, 0, from.Location())
}

// GetPaymentLogsByDateRange retrieves payment logs for a user within a date range
func GetPaymentLogsByDateRange(userID uuid.UUID, startDate, endDate time.Time) ([]PaymentLog, error) {
	query := `
		SELECT id, user_id, recurring_payment_id, amount, payment_date, due_date,
			   payment_method, notes, is_late, late_fee, created_at
		FROM payment_logs
		WHERE user_id = $1 AND payment_date >= $2 AND payment_date <= $3
		ORDER BY payment_date DESC`
	
	rows, err := DB.Query(context.Background(), query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []PaymentLog
	for rows.Next() {
		var log PaymentLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.RecurringPaymentID, &log.Amount,
			&log.PaymentDate, &log.DueDate, &log.PaymentMethod, &log.Notes,
			&log.IsLate, &log.LateFee, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}

// GetPaymentLogs returns all payment logs for a user
func GetPaymentLogs(userID uuid.UUID) ([]PaymentLog, error) {
	query := `
		SELECT id, user_id, recurring_payment_id, amount, payment_date, due_date,
			   payment_method, notes, is_late, late_fee, created_at
		FROM payment_logs
		WHERE user_id = $1
		ORDER BY payment_date DESC`
	
	rows, err := DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []PaymentLog
	for rows.Next() {
		var log PaymentLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.RecurringPaymentID, &log.Amount,
			&log.PaymentDate, &log.DueDate, &log.PaymentMethod, &log.Notes,
			&log.IsLate, &log.LateFee, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}

// GetRecurringPaymentByID retrieves a specific recurring payment by ID
func GetRecurringPaymentByID(userID, paymentID uuid.UUID) (*RecurringPayment, error) {
	query := `
		SELECT id, user_id, name, category, amount, frequency, due_date, start_date,
			   end_date, remaining_payments, total_amount, description, provider,
			   account_number, auto_pay, renewal_notice_days, price_history, metadata,
			   is_active, created_at, updated_at
		FROM recurring_payments
		WHERE user_id = $1 AND id = $2 AND is_active = true`
	
	var payment RecurringPayment
	err := DB.QueryRow(context.Background(), query, userID, paymentID).Scan(
		&payment.ID, &payment.UserID, &payment.Name, &payment.Category,
		&payment.Amount, &payment.Frequency, &payment.DueDate, &payment.StartDate,
		&payment.EndDate, &payment.RemainingPayments, &payment.TotalAmount,
		&payment.Description, &payment.Provider, &payment.AccountNumber,
		&payment.AutoPay, &payment.RenewalNoticeDays, &payment.PriceHistory,
		&payment.Metadata, &payment.IsActive, &payment.CreatedAt, &payment.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, err
	}
	
	// Get user's currency for display
	userFinance, err := GetUserFinance(userID)
	if err == nil {
		payment.Currency = userFinance.Currency
	}
	
	// Calculate additional fields
	calculatePaymentDates(&payment)
	
	return &payment, nil
}

// GetPaymentLogsByRecurringPaymentID retrieves payment logs for a specific recurring payment
func GetPaymentLogsByRecurringPaymentID(userID, recurringPaymentID uuid.UUID) ([]PaymentLog, error) {
	query := `
		SELECT id, user_id, recurring_payment_id, amount, payment_date, due_date,
			   payment_method, notes, is_late, late_fee, created_at
		FROM payment_logs
		WHERE user_id = $1 AND recurring_payment_id = $2
		ORDER BY payment_date DESC`
	
	rows, err := DB.Query(context.Background(), query, userID, recurringPaymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []PaymentLog
	for rows.Next() {
		var log PaymentLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.RecurringPaymentID, &log.Amount,
			&log.PaymentDate, &log.DueDate, &log.PaymentMethod, &log.Notes,
			&log.IsLate, &log.LateFee, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}

// GetPaymentLogByID retrieves a specific payment log by ID
func GetPaymentLogByID(userID, logID uuid.UUID) (*PaymentLog, error) {
	query := `
		SELECT pl.id, pl.user_id, pl.recurring_payment_id, pl.amount, pl.payment_date, pl.due_date,
			   pl.payment_method, pl.notes, pl.is_late, pl.late_fee, pl.created_at,
			   rp.name as payment_name
		FROM payment_logs pl
		LEFT JOIN recurring_payments rp ON pl.recurring_payment_id = rp.id
		WHERE pl.user_id = $1 AND pl.id = $2`
	
	var log PaymentLog
	err := DB.QueryRow(context.Background(), query, userID, logID).Scan(
		&log.ID, &log.UserID, &log.RecurringPaymentID, &log.Amount,
		&log.PaymentDate, &log.DueDate, &log.PaymentMethod, &log.Notes,
		&log.IsLate, &log.LateFee, &log.CreatedAt, &log.PaymentName,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("payment log not found")
		}
		return nil, err
	}
	
	return &log, nil
}

// UpdatePaymentLog updates a payment log
func UpdatePaymentLog(userID, logID uuid.UUID, amount float64, paymentDate, dueDate time.Time, method, notes string) error {
	query := `
		UPDATE payment_logs 
		SET amount = $3, payment_date = $4, due_date = $5, payment_method = $6, notes = $7
		WHERE user_id = $1 AND id = $2`
	
	var methodPtr, notesPtr *string
	if method != "" {
		methodPtr = &method
	}
	if notes != "" {
		notesPtr = &notes
	}
	
	_, err := DB.Exec(context.Background(), query, userID, logID, amount, paymentDate, dueDate, methodPtr, notesPtr)
	return err
}

// DeletePaymentLog deletes a payment log
func DeletePaymentLog(userID, logID uuid.UUID) error {
	query := `DELETE FROM payment_logs WHERE user_id = $1 AND id = $2`
	_, err := DB.Exec(context.Background(), query, userID, logID)
	return err
}
