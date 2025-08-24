package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

// GetFinanceDashboard returns comprehensive financial dashboard data
func GetFinanceDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	dashboard, err := db.GetFinanceDashboard(userID)
	if err != nil {
		http.Error(w, "Failed to get finance dashboard: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// GetUserFinance returns user's financial profile
func GetUserFinance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	finance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to get user finance: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finance)
}

// UpdateUserFinance updates user's financial profile
func UpdateUserFinance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var req struct {
		MonthlyIncome         float64 `json:"monthly_income"`
		EmergencyBufferAmount float64 `json:"emergency_buffer_amount"`
		EmergencyBufferPercent float64 `json:"emergency_buffer_percentage"`
		Currency              string  `json:"currency"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate input
	if req.MonthlyIncome < 0 {
		http.Error(w, "Monthly income cannot be negative", http.StatusBadRequest)
		return
	}
	
	if req.EmergencyBufferPercent < 0 || req.EmergencyBufferPercent > 100 {
		http.Error(w, "Emergency buffer percentage must be between 0 and 100", http.StatusBadRequest)
		return
	}
	
	if req.Currency == "" {
		req.Currency = "USD"
	}
	
	err := db.UpdateUserFinance(userID, req.MonthlyIncome, req.EmergencyBufferAmount, req.EmergencyBufferPercent, req.Currency)
	if err != nil {
		http.Error(w, "Failed to update user finance: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GetRecurringPayments returns all recurring payments for the user
func GetRecurringPayments(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	payments, err := db.GetRecurringPayments(userID)
	if err != nil {
		http.Error(w, "Failed to get recurring payments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

// CreateRecurringPayment creates a new recurring payment
func CreateRecurringPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var payment db.RecurringPayment
	
	// Handle both JSON and form data
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// JSON request
		if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		payment.Name = r.FormValue("name")
		payment.Category = r.FormValue("category")
		payment.Frequency = r.FormValue("frequency")
		if desc := r.FormValue("description"); desc != "" {
			payment.Description = &desc
		}
		payment.AutoPay = r.FormValue("auto_pay") == "on"
		
		// Parse amount
		if amountStr := r.FormValue("amount"); amountStr != "" {
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				http.Error(w, "Invalid amount", http.StatusBadRequest)
				return
			}
			payment.Amount = amount
		}
		
		// Parse start date
		if startDateStr := r.FormValue("start_date"); startDateStr != "" {
			startDate, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				http.Error(w, "Invalid start date", http.StatusBadRequest)
				return
			}
			payment.StartDate = startDate
		}
		
		// Parse due day of month
		if dueDayStr := r.FormValue("due_day"); dueDayStr != "" {
			dueDay, err := strconv.Atoi(dueDayStr)
			if err != nil || dueDay < 1 || dueDay > 31 {
				http.Error(w, "Invalid due day - must be between 1 and 31", http.StatusBadRequest)
				return
			}
			payment.DueDate = dueDay
		}
		
		// Parse end date if provided
		if endDateStr := r.FormValue("end_date"); endDateStr != "" {
			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				http.Error(w, "Invalid end date", http.StatusBadRequest)
				return
			}
			payment.EndDate = &endDate
		}
		
		// Parse remaining payments
		if remainingStr := r.FormValue("remaining_payments"); remainingStr != "" {
			remaining, err := strconv.Atoi(remainingStr)
			if err != nil {
				http.Error(w, "Invalid remaining payments", http.StatusBadRequest)
				return
			}
			payment.RemainingPayments = &remaining
		}
		
		// Parse total amount for loans
		if totalAmountStr := r.FormValue("total_amount"); totalAmountStr != "" {
			totalAmount, err := strconv.ParseFloat(totalAmountStr, 64)
			if err != nil {
				http.Error(w, "Invalid total amount", http.StatusBadRequest)
				return
			}
			payment.TotalAmount = &totalAmount
		}
		
		// Parse provider and account number
		if provider := r.FormValue("provider"); provider != "" {
			payment.Provider = &provider
		}
		
		if accountNumber := r.FormValue("account_number"); accountNumber != "" {
			payment.AccountNumber = &accountNumber
		}
	}
	
	// Validate required fields
	if payment.Name == "" {
		http.Error(w, "Payment name is required", http.StatusBadRequest)
		return
	}
	
	if payment.Amount <= 0 {
		http.Error(w, "Payment amount must be positive", http.StatusBadRequest)
		return
	}
	
	if payment.Category == "" {
		http.Error(w, "Payment category is required", http.StatusBadRequest)
		return
	}
	
	if payment.Frequency == "" {
		http.Error(w, "Payment frequency is required", http.StatusBadRequest)
		return
	}
	
	if payment.DueDate < 1 || payment.DueDate > 31 {
		http.Error(w, "Due date must be between 1 and 31", http.StatusBadRequest)
		return
	}
	
	// Set user ID and defaults
	payment.UserID = userID
	if payment.RenewalNoticeDays == 0 {
		payment.RenewalNoticeDays = 30
	}
	
	// Parse start date if provided as string
	if payment.StartDate.IsZero() {
		payment.StartDate = time.Now()
	}
	
	err := db.CreateRecurringPayment(&payment)
	if err != nil {
		http.Error(w, "Failed to create recurring payment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payment)
}

// UpdateRecurringPayment updates an existing recurring payment
func UpdateRecurringPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	vars := mux.Vars(r)
	paymentID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}
	
	// Parse form data
	var payment db.RecurringPayment
	payment.ID = paymentID
	payment.UserID = userID
	
	// Parse basic fields
	payment.Name = r.FormValue("name")
	payment.Category = r.FormValue("category")
	payment.Frequency = r.FormValue("frequency")
	
	// Parse optional description
	if description := r.FormValue("description"); description != "" {
		payment.Description = &description
	}
	
	// Parse amount
	amountStr := r.FormValue("amount")
	if amountStr != "" {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			http.Error(w, "Invalid amount", http.StatusBadRequest)
			return
		}
		payment.Amount = amount
	}
	
	// Parse due date  
	dueDateStr := r.FormValue("due_date")
	if dueDateStr != "" {
		dueDate, err := strconv.Atoi(dueDateStr)
		if err != nil {
			http.Error(w, "Invalid due date", http.StatusBadRequest)
			return
		}
		payment.DueDate = dueDate
	}
	
	// Parse auto pay
	payment.AutoPay = r.FormValue("auto_pay") == "true"
	
	// Parse optional dates
	if startDateStr := r.FormValue("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start date", http.StatusBadRequest)
			return
		}
		payment.StartDate = startDate
	}
	
	if endDateStr := r.FormValue("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end date", http.StatusBadRequest)
			return
		}
		payment.EndDate = &endDate
	}
	
	// Parse remaining payments
	if remainingStr := r.FormValue("remaining_payments"); remainingStr != "" {
		remaining, err := strconv.Atoi(remainingStr)
		if err != nil {
			http.Error(w, "Invalid remaining payments", http.StatusBadRequest)
			return
		}
		payment.RemainingPayments = &remaining
	}
	
	// Parse total amount for loans
	if totalAmountStr := r.FormValue("total_amount"); totalAmountStr != "" {
		totalAmount, err := strconv.ParseFloat(totalAmountStr, 64)
		if err != nil {
			http.Error(w, "Invalid total amount", http.StatusBadRequest)
			return
		}
		payment.TotalAmount = &totalAmount
	}
	
	// Parse provider and account number
	if provider := r.FormValue("provider"); provider != "" {
		payment.Provider = &provider
	}
	
	if accountNumber := r.FormValue("account_number"); accountNumber != "" {
		payment.AccountNumber = &accountNumber
	}
	
	// Validate required fields
	if payment.Name == "" {
		http.Error(w, "Payment name is required", http.StatusBadRequest)
		return
	}
	
	if payment.Amount <= 0 {
		http.Error(w, "Payment amount must be positive", http.StatusBadRequest)
		return
	}
	
	err = db.UpdateRecurringPayment(&payment)
	if err != nil {
		http.Error(w, "Failed to update recurring payment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// DeleteRecurringPayment soft deletes a recurring payment
func DeleteRecurringPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	vars := mux.Vars(r)
	paymentID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}
	
	err = db.DeleteRecurringPayment(paymentID, userID)
	if err != nil {
		http.Error(w, "Failed to delete recurring payment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GetSavingsGoals returns all savings goals for the user
func GetSavingsGoals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	goals, err := db.GetSavingsGoals(userID)
	if err != nil {
		http.Error(w, "Failed to get savings goals: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

// CreateSavingsGoal creates a new savings goal
func CreateSavingsGoal(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var goal db.SavingsGoal
	
	// Handle both JSON and form data
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// JSON request
		if err := json.NewDecoder(r.Body).Decode(&goal); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		goal.Name = r.FormValue("name")
		goal.Priority = r.FormValue("priority")
		if desc := r.FormValue("description"); desc != "" {
			goal.Description = &desc
		}
		if productURL := r.FormValue("product_link"); productURL != "" {
			goal.ProductURL = &productURL
		}
		if imageURL := r.FormValue("image_url"); imageURL != "" {
			goal.ImageURL = &imageURL
		}
		
		// Parse target_amount
		if targetAmountStr := r.FormValue("target_amount"); targetAmountStr != "" {
			targetAmount, err := strconv.ParseFloat(targetAmountStr, 64)
			if err != nil {
				http.Error(w, "Invalid target amount", http.StatusBadRequest)
				return
			}
			goal.TargetAmount = targetAmount
		}
		
		// Parse monthly_contribution
		if monthlyContribStr := r.FormValue("monthly_contribution"); monthlyContribStr != "" {
			monthlyContrib, err := strconv.ParseFloat(monthlyContribStr, 64)
			if err != nil {
				http.Error(w, "Invalid monthly contribution", http.StatusBadRequest)
				return
			}
			goal.MonthlyContribution = monthlyContrib
		}
		
		// Calculate target date from target_months
		if targetMonthsStr := r.FormValue("target_months"); targetMonthsStr != "" {
			targetMonths, err := strconv.Atoi(targetMonthsStr)
			if err != nil {
				http.Error(w, "Invalid target months", http.StatusBadRequest)
				return
			}
			targetDate := time.Now().AddDate(0, targetMonths, 0)
			goal.TargetDate = &targetDate
		}
	}
	
	// Validate required fields
	if goal.Name == "" {
		http.Error(w, "Goal name is required", http.StatusBadRequest)
		return
	}
	
	if goal.TargetAmount <= 0 {
		http.Error(w, "Target amount must be positive", http.StatusBadRequest)
		return
	}
	
	if goal.Priority == "" {
		goal.Priority = "nice_to_have"
	}
	
	// Set user ID and defaults
	goal.UserID = userID
	
	err := db.CreateSavingsGoal(&goal)
	if err != nil {
		http.Error(w, "Failed to create savings goal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(goal)
}

// UpdateSavingsGoal updates an existing savings goal
func UpdateSavingsGoal(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	vars := mux.Vars(r)
	goalID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}
	
	var goal db.SavingsGoal
	if err := json.NewDecoder(r.Body).Decode(&goal); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Set IDs
	goal.ID = goalID
	goal.UserID = userID
	
	// Validate required fields
	if goal.Name == "" {
		http.Error(w, "Goal name is required", http.StatusBadRequest)
		return
	}
	
	if goal.TargetAmount <= 0 {
		http.Error(w, "Target amount must be positive", http.StatusBadRequest)
		return
	}
	
	err = db.UpdateSavingsGoal(&goal)
	if err != nil {
		http.Error(w, "Failed to update savings goal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// AddSavingsContribution adds money to a savings goal
func AddSavingsContribution(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	vars := mux.Vars(r)
	goalID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Parse form data
	amountStr := r.FormValue("amount")
	source := r.FormValue("source")
	notes := r.FormValue("notes")
	contributionDateStr := r.FormValue("contribution_date")
	
	// Validate required fields
	if amountStr == "" {
		http.Error(w, "Contribution amount is required", http.StatusBadRequest)
		return
	}
	
	// Parse amount
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid contribution amount", http.StatusBadRequest)
		return
	}
	
	if amount <= 0 {
		http.Error(w, "Contribution amount must be positive", http.StatusBadRequest)
		return
	}
	
	// Parse contribution date (optional)
	var contributionDate time.Time
	if contributionDateStr != "" {
		contributionDate, err = time.Parse("2006-01-02", contributionDateStr)
		if err != nil {
			http.Error(w, "Invalid contribution date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		contributionDate = time.Now()
	}
	
	if source == "" {
		source = "manual"
	}
	
	err = db.AddSavingsContributionWithDate(userID, goalID, amount, contributionDate, source, notes)
	if err != nil {
		http.Error(w, "Failed to add contribution: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// For HTMX, return a simple success message that will replace the modal
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="success-message">Contribution added successfully!</div>`))
}

// AddSavingsContributionFromForm handles form-based savings contributions
func AddSavingsContributionFromForm(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse form data
	goalIDStr := r.FormValue("savings_goal_id")
	amountStr := r.FormValue("amount")
	source := r.FormValue("source")
	notes := r.FormValue("notes")
	contributionDateStr := r.FormValue("contribution_date")

	// Validate required fields
	if goalIDStr == "" {
		http.Error(w, "Savings goal selection is required", http.StatusBadRequest)
		return
	}

	if amountStr == "" {
		http.Error(w, "Contribution amount is required", http.StatusBadRequest)
		return
	}

	// Parse goal ID
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		http.Error(w, "Invalid savings goal ID", http.StatusBadRequest)
		return
	}

	// Parse amount
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid contribution amount", http.StatusBadRequest)
		return
	}

	if amount <= 0 {
		http.Error(w, "Contribution amount must be positive", http.StatusBadRequest)
		return
	}

	// Parse contribution date (optional)
	var contributionDate time.Time
	if contributionDateStr != "" {
		contributionDate, err = time.Parse("2006-01-02", contributionDateStr)
		if err != nil {
			http.Error(w, "Invalid contribution date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		contributionDate = time.Now()
	}

	if source == "" {
		source = "manual"
	}

	err = db.AddSavingsContributionWithDate(userID, goalID, amount, contributionDate, source, notes)
	if err != nil {
		http.Error(w, "Failed to add contribution: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return HTML that triggers JavaScript for HTMX
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<script>showToast('Contribution added successfully!', 'success');</script>`))
}

// LogPayment records a payment for a recurring payment
func LogPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Parse form data
	recurringPaymentIDStr := r.FormValue("recurring_payment_id")
	amountStr := r.FormValue("amount_paid") // Form sends as amount_paid
	paymentDateStr := r.FormValue("payment_date")
	methodStr := r.FormValue("payment_method")
	notesStr := r.FormValue("notes")
	
	// Validate required fields
	if recurringPaymentIDStr == "" {
		http.Error(w, "Recurring payment ID is required", http.StatusBadRequest)
		return
	}
	
	if amountStr == "" {
		http.Error(w, "Payment amount is required", http.StatusBadRequest)
		return
	}
	
	if paymentDateStr == "" {
		http.Error(w, "Payment date is required", http.StatusBadRequest)
		return
	}
	
	// Parse amount
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid payment amount", http.StatusBadRequest)
		return
	}
	
	if amount <= 0 {
		http.Error(w, "Payment amount must be positive", http.StatusBadRequest)
		return
	}
	
	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", paymentDateStr)
	if err != nil {
		http.Error(w, "Invalid payment date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	
	// Parse recurring payment ID
	recurringPaymentID, err := uuid.Parse(recurringPaymentIDStr)
	if err != nil {
		http.Error(w, "Invalid recurring payment ID", http.StatusBadRequest)
		return
	}
	
	// Get the recurring payment to determine due date
	recurringPayment, err := db.GetRecurringPaymentByID(userID, recurringPaymentID)
	if err != nil {
		http.Error(w, "Failed to find recurring payment: "+err.Error(), http.StatusNotFound)
		return
	}
	
	// For the due date, use the recurring payment's next due date or the payment date if it's in advance
	dueDate := paymentDate
	if recurringPayment.NextDueDate != nil {
		// If payment is for current month, use the next due date
		currentTime := time.Now()
		if paymentDate.Month() == currentTime.Month() && paymentDate.Year() == currentTime.Year() {
			dueDate = *recurringPayment.NextDueDate
		}
		// For advance payments (future months), use the payment date as due date
	}
	
	err = db.LogPayment(userID, &recurringPaymentID, amount, paymentDate, dueDate, methodStr, notesStr)
	if err != nil {
		http.Error(w, "Failed to log payment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// For HTMX, return a simple success message that will replace the modal
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="success-message">Payment logged successfully!</div>`))
}

// CalculateSafeToSpend returns the current safe-to-spend amount
func CalculateSafeToSpend(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	dashboard, err := db.GetFinanceDashboard(userID)
	if err != nil {
		http.Error(w, "Failed to calculate safe to spend: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	result := map[string]interface{}{
		"safe_to_spend":        dashboard.SafeToSpend,
		"monthly_income":       dashboard.UserFinance.MonthlyIncome,
		"monthly_commitments":  dashboard.MonthlyCommitments,
		"emergency_buffer":     dashboard.UserFinance.EmergencyBufferAmount,
		"total_savings_goals":  func() float64 {
			var total float64
			for _, goal := range dashboard.ActiveSavingsGoals {
				total += goal.MonthlyContribution
			}
			return total
		}(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetPaymentReminders returns upcoming payment reminders
func GetPaymentReminders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get days parameter (default to 7 days)
	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}
	
	payments, err := db.GetRecurringPayments(userID)
	if err != nil {
		http.Error(w, "Failed to get payment reminders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Filter payments due within the specified days
	now := time.Now()
	cutoff := now.AddDate(0, 0, days)
	
	var reminders []db.RecurringPayment
	for _, payment := range payments {
		if payment.NextDueDate != nil && payment.NextDueDate.Before(cutoff) {
			reminders = append(reminders, payment)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminders)
}