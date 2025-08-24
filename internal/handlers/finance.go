package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

// FinanceHandler renders the main finance dashboard page
func FinanceHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("FinanceHandler called - Method: %s, Headers: %v", r.Method, r.Header)
	
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		log.Printf("FinanceHandler: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("FinanceHandler: userID = %s", userID)
	
	log.Printf("FinanceHandler: Getting dashboard for user %s", userID)
	dashboard, err := db.GetFinanceDashboard(userID)
	if err != nil {
		log.Printf("FinanceHandler: Error getting dashboard: %v", err)
		http.Error(w, "Failed to load finance dashboard: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("FinanceHandler: Dashboard retrieved successfully")
	
	// Get user data for header display
	user, err := db.GetUserByID(db.DB, userID)
	if err != nil {
		log.Printf("FinanceHandler: Error getting user: %v", err)
		http.Error(w, "Failed to load user data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Dashboard": dashboard,
		"User":      user,
	}
	
	// Create a new template with the same functions as the global template
	log.Printf("FinanceHandler: Creating new template with custom functions")
	financeTmpl := template.New("").Funcs(template.FuncMap{
		"now": time.Now,
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"value": func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		},
		"deref": func(p *float64) float64 {
			if p == nil {
				return 0.0
			}
			return *p
		},
	})
	
	log.Printf("FinanceHandler: Parsing template files")
	financeTmpl, err = financeTmpl.ParseFiles(
		"templates/base.html",
		"templates/finance.html",
		"templates/partials/finance_dashboard.html",
		"templates/partials/payment_item.html",
		"templates/partials/savings_goal_item.html",
	)
	if err != nil {
		log.Printf("FinanceHandler: Error parsing templates: %v", err)
		http.Error(w, "Failed to parse templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("FinanceHandler: Templates parsed successfully")
	
	// Use a buffer to execute template first, then write to response if successful
	log.Printf("FinanceHandler: Executing template")
	var buf bytes.Buffer
	err = financeTmpl.ExecuteTemplate(&buf, "base.html", data)
	if err != nil {
		log.Printf("FinanceHandler: Error executing template: %v", err)
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("FinanceHandler: Template executed successfully")
	
	// Only write to response if template execution was successful
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

// FinanceSetupHandler renders the financial setup/settings page
func FinanceSetupHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load finance profile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get user data for header display
	user, err := db.GetUserByID(db.DB, userID)
	if err != nil {
		log.Printf("FinanceSetupHandler: Error getting user: %v", err)
		http.Error(w, "Failed to load user data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"UserFinance": userFinance,
		"User":        user,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/finance_setup.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateFinanceProfileHandler updates user's financial profile
func UpdateFinanceProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	monthlyIncome, err := strconv.ParseFloat(r.FormValue("monthly_income"), 64)
	if err != nil {
		http.Error(w, "Invalid monthly income", http.StatusBadRequest)
		return
	}
	
	emergencyBufferAmount, err := strconv.ParseFloat(r.FormValue("emergency_buffer_amount"), 64)
	if err != nil {
		emergencyBufferAmount = 0 // Optional field
	}
	
	emergencyBufferPercent, err := strconv.ParseFloat(r.FormValue("emergency_buffer_percentage"), 64)
	if err != nil {
		emergencyBufferPercent = 10.0 // Default
	}
	
	currency := r.FormValue("currency")
	if currency == "" {
		currency = "USD"
	}
	
	err = db.UpdateUserFinance(userID, monthlyIncome, emergencyBufferAmount, emergencyBufferPercent, currency)
	if err != nil {
		http.Error(w, "Failed to update finance profile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Redirect back to finance page
	http.Redirect(w, r, "/finance", http.StatusSeeOther)
}

// PaymentsListHandler renders the recurring payments list
func PaymentsListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	payments, err := db.GetRecurringPayments(userID)
	if err != nil {
		http.Error(w, "Failed to load payments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Payments": payments,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/payments_list.html",
		"templates/partials/payment_item.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "payments_list.html", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// AddPaymentHandler handles adding a new recurring payment
func AddPaymentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	payment := db.RecurringPayment{
		UserID:      userID,
		Name:        r.FormValue("name"),
		Category:    r.FormValue("category"),
		Description: stringPtr(r.FormValue("description")),
		Provider:    stringPtr(r.FormValue("provider")),
		Frequency:   r.FormValue("frequency"),
	}
	
	// Parse amount
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	payment.Amount = amount
	
	// Parse due date
	dueDate, err := strconv.Atoi(r.FormValue("due_date"))
	if err != nil {
		http.Error(w, "Invalid due date", http.StatusBadRequest)
		return
	}
	payment.DueDate = dueDate
	
	// Parse start date
	startDateStr := r.FormValue("start_date")
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start date format", http.StatusBadRequest)
			return
		}
		payment.StartDate = startDate
	} else {
		payment.StartDate = time.Now()
	}
	
	// Parse end date (optional)
	endDateStr := r.FormValue("end_date")
	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end date format", http.StatusBadRequest)
			return
		}
		payment.EndDate = &endDate
	}
	
	// Parse remaining payments (optional, for loans)
	remainingStr := r.FormValue("remaining_payments")
	if remainingStr != "" {
		remaining, err := strconv.Atoi(remainingStr)
		if err == nil {
			payment.RemainingPayments = &remaining
		}
	}
	
	// Parse total amount (optional, for loans)
	totalStr := r.FormValue("total_amount")
	if totalStr != "" {
		total, err := strconv.ParseFloat(totalStr, 64)
		if err == nil {
			payment.TotalAmount = &total
		}
	}
	
	// Parse auto pay
	payment.AutoPay = r.FormValue("auto_pay") == "on"
	
	err = db.CreateRecurringPayment(&payment)
	if err != nil {
		http.Error(w, "Failed to create payment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return updated payments list
	PaymentsListHandler(w, r)
}

// EditPaymentFormHandler renders the edit payment form
func EditPaymentFormHandler(w http.ResponseWriter, r *http.Request) {
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
	
	// Get the payment (this would need a function in db package)
	payments, err := db.GetRecurringPayments(userID)
	if err != nil {
		http.Error(w, "Failed to load payment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	var payment *db.RecurringPayment
	for _, p := range payments {
		if p.ID == paymentID {
			payment = &p
			break
		}
	}
	
	if payment == nil {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}
	
	data := map[string]interface{}{
		"Payment": payment,
	}
	
	tmpl := template.New("").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"value": func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		},
		"deref": func(p *float64) float64 {
			if p == nil {
				return 0.0
			}
			return *p
		},
	})
	tmpl = template.Must(tmpl.ParseFiles(
		"templates/partials/payment_edit_form.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "payment_edit_form", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeletePaymentHandler handles deleting a recurring payment
func DeletePaymentHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Failed to delete payment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return success response for HTMX
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Payment deleted successfully"}`))
}

// AllPaymentsHandler shows all recurring payments in a modal
func AllPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get all recurring payments
	payments, err := db.GetRecurringPayments(userID)
	if err != nil {
		http.Error(w, "Failed to load payments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get user finance info for currency
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load user finance info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Payments":    payments,
		"UserFinance": userFinance,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/all_payments_modal.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "all_payments_modal", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// SavingsGoalsListHandler renders the savings goals list
func SavingsGoalsListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	goals, err := db.GetSavingsGoals(userID)
	if err != nil {
		http.Error(w, "Failed to load savings goals: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Goals": goals,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/savings_goals_list.html",
		"templates/partials/savings_goal_item.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "savings_goals_list.html", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// AddSavingsGoalHandler handles adding a new savings goal
func AddSavingsGoalHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	goal := db.SavingsGoal{
		UserID:      userID,
		Name:        r.FormValue("name"),
		Priority:    r.FormValue("priority"),
		Description: stringPtr(r.FormValue("description")),
		ProductURL:  stringPtr(r.FormValue("product_url")),
		ImageURL:    stringPtr(r.FormValue("image_url")),
	}
	
	// Parse target amount
	targetAmount, err := strconv.ParseFloat(r.FormValue("target_amount"), 64)
	if err != nil {
		http.Error(w, "Invalid target amount", http.StatusBadRequest)
		return
	}
	goal.TargetAmount = targetAmount
	
	// Parse monthly contribution
	monthlyContrib, err := strconv.ParseFloat(r.FormValue("monthly_contribution"), 64)
	if err == nil {
		goal.MonthlyContribution = monthlyContrib
	}
	
	// Parse target months and calculate target date
	targetMonthsStr := r.FormValue("target_months")
	if targetMonthsStr != "" {
		targetMonths, err := strconv.Atoi(targetMonthsStr)
		if err != nil {
			http.Error(w, "Invalid target months", http.StatusBadRequest)
			return
		}
		if targetMonths > 0 {
			targetDate := time.Now().AddDate(0, targetMonths, 0)
			goal.TargetDate = &targetDate
		}
	}
	
	// Auto calculate contribution if enabled
	goal.AutoCalculateContrib = r.FormValue("auto_calculate") == "on"
	
	err = db.CreateSavingsGoal(&goal)
	if err != nil {
		http.Error(w, "Failed to create savings goal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return updated finance dashboard
	FinanceHandler(w, r)
}

// AddContributionHandler handles adding money to a savings goal
func AddContributionHandler(w http.ResponseWriter, r *http.Request) {
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
	
	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	
	method := r.FormValue("method")
	if method == "" {
		method = "manual"
	}
	
	notes := r.FormValue("notes")
	
	err = db.AddSavingsContribution(userID, goalID, amount, method, notes)
	if err != nil {
		http.Error(w, "Failed to add contribution: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return updated finance dashboard
	FinanceHandler(w, r)
}

// QuickAddContributionHandler handles quick contribution from the dashboard
func QuickAddContributionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Debug: Log all form values
	log.Printf("QuickAddContributionHandler: Form values: %v", r.Form)
	log.Printf("QuickAddContributionHandler: goal_id = '%s'", r.FormValue("goal_id"))
	log.Printf("QuickAddContributionHandler: amount = '%s'", r.FormValue("amount"))
	
	// Get goal ID from select field
	goalIDStr := r.FormValue("goal_id")
	if goalIDStr == "" {
		log.Printf("QuickAddContributionHandler: No goal ID found")
		http.Error(w, "Please select a savings goal", http.StatusBadRequest)
		return
	}
	
	goalID, err := uuid.Parse(goalIDStr)
	if err != nil {
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}
	
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	
	source := r.FormValue("source")
	if source == "" {
		source = "manual"
	}
	
	notes := r.FormValue("notes")
	
	err = db.AddSavingsContribution(userID, goalID, amount, source, notes)
	if err != nil {
		http.Error(w, "Failed to add contribution: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return updated finance dashboard
	FinanceHandler(w, r)
}

// LogPaymentHandler handles logging a payment
func LogPaymentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	
	// Parse payment date
	paymentDateStr := r.FormValue("payment_date")
	if paymentDateStr == "" {
		paymentDateStr = time.Now().Format("2006-01-02")
	}
	paymentDate, err := time.Parse("2006-01-02", paymentDateStr)
	if err != nil {
		http.Error(w, "Invalid payment date", http.StatusBadRequest)
		return
	}
	
	// Parse due date
	dueDateStr := r.FormValue("due_date")
	dueDate, err := time.Parse("2006-01-02", dueDateStr)
	if err != nil {
		http.Error(w, "Invalid due date", http.StatusBadRequest)
		return
	}
	
	// Parse recurring payment ID (optional)
	var recurringPaymentID *uuid.UUID
	recurringIDStr := r.FormValue("recurring_payment_id")
	if recurringIDStr != "" {
		id, err := uuid.Parse(recurringIDStr)
		if err == nil {
			recurringPaymentID = &id
		}
	}
	
	method := r.FormValue("method")
	notes := r.FormValue("notes")
	
	err = db.LogPayment(userID, recurringPaymentID, amount, paymentDate, dueDate, method, notes)
	if err != nil {
		http.Error(w, "Failed to log payment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Redirect to finance dashboard
	http.Redirect(w, r, "/finance", http.StatusSeeOther)
}

// PaymentHistoryHandler shows payment history for a recurring payment
func PaymentHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get payment ID from URL
	vars := mux.Vars(r)
	paymentIDStr := vars["id"]
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}
	
	// Get the recurring payment
	payment, err := db.GetRecurringPaymentByID(userID, paymentID)
	if err != nil {
		http.Error(w, "Payment not found: "+err.Error(), http.StatusNotFound)
		return
	}
	
	// Get payment logs for this recurring payment
	logs, err := db.GetPaymentLogsByRecurringPaymentID(userID, paymentID)
	if err != nil {
		http.Error(w, "Failed to get payment logs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Payment": payment,
		"PaymentLogs": logs,
	}
	
	// Create template with custom functions
	paymentHistoryTmpl := template.New("").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("Jan 2, 2006 3:04 PM")
		},
	})
	
	paymentHistoryTmpl, err = paymentHistoryTmpl.ParseFiles("templates/partials/payment_history_modal.html")
	if err != nil {
		log.Printf("Error parsing payment history template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	
	err = paymentHistoryTmpl.ExecuteTemplate(w, "payment_history_modal", data)
	if err != nil {
		log.Printf("Error executing payment history template: %v", err)
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// EditPaymentLogHandler shows form to edit a payment log
func EditPaymentLogHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get payment log ID from URL
	vars := mux.Vars(r)
	logIDStr := vars["id"]
	logID, err := uuid.Parse(logIDStr)
	if err != nil {
		http.Error(w, "Invalid payment log ID", http.StatusBadRequest)
		return
	}
	
	// Get the payment log
	paymentLog, err := db.GetPaymentLogByID(userID, logID)
	if err != nil {
		http.Error(w, "Payment log not found: "+err.Error(), http.StatusNotFound)
		return
	}
	
	// Create safe template data with dereferenced pointers
	safePaymentLog := struct {
		ID                  string
		RecurringPaymentID  *uuid.UUID
		Amount              float64
		PaymentDate         time.Time
		DueDate             time.Time
		PaymentMethod       string
		Notes               string
		PaymentName         string
	}{
		ID:                  paymentLog.ID.String(),
		RecurringPaymentID:  paymentLog.RecurringPaymentID,
		Amount:              paymentLog.Amount,
		PaymentDate:         paymentLog.PaymentDate,
		DueDate:             paymentLog.DueDate,
		PaymentMethod:       "",
		Notes:               "",
		PaymentName:         paymentLog.PaymentName,
	}
	
	// Safely dereference pointer fields
	if paymentLog.PaymentMethod != nil {
		safePaymentLog.PaymentMethod = *paymentLog.PaymentMethod
	}
	if paymentLog.Notes != nil {
		safePaymentLog.Notes = *paymentLog.Notes
	}
	
	data := map[string]interface{}{
		"PaymentLog": safePaymentLog,
	}
	
	// Create template with custom functions
	editLogTmpl := template.New("").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	})
	
	editLogTmpl, err = editLogTmpl.ParseFiles("templates/partials/payment_log_edit_form.html")
	if err != nil {
		log.Printf("Error parsing edit payment log template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	
	err = editLogTmpl.ExecuteTemplate(w, "payment_log_edit_form", data)
	if err != nil {
		log.Printf("Error executing edit payment log template: %v", err)
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// UpdatePaymentLogHandler updates a payment log
func UpdatePaymentLogHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get payment log ID from URL
	vars := mux.Vars(r)
	logIDStr := vars["id"]
	logID, err := uuid.Parse(logIDStr)
	if err != nil {
		http.Error(w, "Invalid payment log ID", http.StatusBadRequest)
		return
	}
	
	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Parse amount
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	
	// Parse payment date
	paymentDateStr := r.FormValue("payment_date")
	paymentDate, err := time.Parse("2006-01-02", paymentDateStr)
	if err != nil {
		http.Error(w, "Invalid payment date", http.StatusBadRequest)
		return
	}
	
	// Parse due date
	dueDateStr := r.FormValue("due_date")
	dueDate, err := time.Parse("2006-01-02", dueDateStr)
	if err != nil {
		http.Error(w, "Invalid due date", http.StatusBadRequest)
		return
	}
	
	method := r.FormValue("method")
	notes := r.FormValue("notes")
	
	err = db.UpdatePaymentLog(userID, logID, amount, paymentDate, dueDate, method, notes)
	if err != nil {
		http.Error(w, "Failed to update payment log: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return success response for HTMX
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Payment log updated successfully"}`))
}

// DeletePaymentLogHandler deletes a payment log
func DeletePaymentLogHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get payment log ID from URL
	vars := mux.Vars(r)
	logIDStr := vars["id"]
	logID, err := uuid.Parse(logIDStr)
	if err != nil {
		http.Error(w, "Invalid payment log ID", http.StatusBadRequest)
		return
	}
	
	err = db.DeletePaymentLog(userID, logID)
	if err != nil {
		http.Error(w, "Failed to delete payment log: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return empty response for HTMX to remove the element
	w.WriteHeader(http.StatusOK)
}

// GetSafeToSpendHandler returns just the safe-to-spend amount as HTML
func GetSafeToSpendHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	dashboard, err := db.GetFinanceDashboard(userID)
	if err != nil {
		http.Error(w, "Failed to calculate safe to spend", http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"SafeToSpend": dashboard.SafeToSpend,
		"Currency":    dashboard.UserFinance.Currency,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/safe_to_spend.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "safe_to_spend.html", data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// QuickActionsHandler renders quick action buttons/forms
func QuickActionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get recent payments for quick logging
	payments, err := db.GetRecurringPayments(userID)
	if err != nil {
		http.Error(w, "Failed to load payments", http.StatusInternalServerError)
		return
	}
	
	// Get active goals for quick contributions
	goals, err := db.GetSavingsGoals(userID)
	if err != nil {
		http.Error(w, "Failed to load goals", http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Payments": payments,
		"Goals":    goals,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/finance_quick_actions.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "finance_quick_actions.html", data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// NewPaymentFormHandler renders the form to add a new recurring payment
func NewPaymentFormHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	data := map[string]interface{}{
		"UserID": userID,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/new_payment_form.html",
	))
	
	err := tmpl.ExecuteTemplate(w, "new_payment_form", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// NewSavingsGoalFormHandler renders the form to add a new savings goal
func NewSavingsGoalFormHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get user finance info for currency
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load user finance info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"UserID": userID,
		"UserFinance": userFinance,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/new_savings_goal_form.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "new_savings_goal_form", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// QuickPaymentFormHandler renders the form for quick payment logging
func QuickPaymentFormHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	payments, err := db.GetRecurringPayments(userID)
	if err != nil {
		http.Error(w, "Failed to load payments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Payments": payments,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/quick_payment_form.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "quick_payment_form", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// QuickContributionFormHandler renders the form for quick savings contributions
func QuickContributionFormHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	goals, err := db.GetSavingsGoals(userID)
	if err != nil {
		http.Error(w, "Failed to load savings goals: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"SavingsGoals": goals,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/quick_contribution_form.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "quick_contribution_form", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// FinanceAnalyticsHandler renders the finance analytics page
func FinanceAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	dashboard, err := db.GetFinanceDashboard(userID)
	if err != nil {
		http.Error(w, "Failed to load finance dashboard: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get user data for header display
	user, err := db.GetUserByID(db.DB, userID)
	if err != nil {
		log.Printf("FinanceAnalyticsHandler: Error getting user: %v", err)
		http.Error(w, "Failed to load user data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Calculate analytics percentages
	analytics := make(map[string]interface{})
	monthlyIncome := dashboard.UserFinance.MonthlyIncome
	
	if monthlyIncome > 0 {
		analytics["SavingsRatePercent"] = (dashboard.TotalSavingsContrib / monthlyIncome) * 100
		analytics["FixedPaymentsPercent"] = (dashboard.MonthlyCommitments / monthlyIncome) * 100
		analytics["SavingsGoalsPercent"] = (dashboard.TotalSavingsContrib / monthlyIncome) * 100
		analytics["EmergencyBufferPercent"] = (dashboard.CalculatedEmergencyBuffer / monthlyIncome) * 100
		analytics["SafeToSpendPercent"] = (dashboard.SafeToSpend / monthlyIncome) * 100
	} else {
		analytics["SavingsRatePercent"] = nil
		analytics["FixedPaymentsPercent"] = nil
		analytics["SavingsGoalsPercent"] = nil
		analytics["EmergencyBufferPercent"] = nil
		analytics["SafeToSpendPercent"] = nil
	}
	
	data := map[string]interface{}{
		"Dashboard": dashboard,
		"User":      user,
		"Analytics": analytics,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/finance_analytics.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// QuickPayFormHandler renders the form for quick payment of a specific recurring payment
func QuickPayFormHandler(w http.ResponseWriter, r *http.Request) {
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
	
	payments, err := db.GetRecurringPayments(userID)
	if err != nil {
		http.Error(w, "Failed to load payments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Find the specific payment
	var payment *db.RecurringPayment
	for _, p := range payments {
		if p.ID == paymentID {
			payment = &p
			break
		}
	}
	
	if payment == nil {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}
	
	data := map[string]interface{}{
		"Payment": payment,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/quick_pay_form.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "quick_pay_form", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// EditSavingsGoalFormHandler shows the edit form for a savings goal
func EditSavingsGoalFormHandler(w http.ResponseWriter, r *http.Request) {
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
	
	goal, err := db.GetSavingsGoalByID(goalID, userID)
	if err != nil {
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}
	
	// Get user finance info for currency
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load user finance info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Goal": goal,
		"UserFinance": userFinance,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/edit_savings_goal_form.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "edit_savings_goal_form", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// SavingsGoalContributeFormHandler shows the contribution form for a savings goal
func SavingsGoalContributeFormHandler(w http.ResponseWriter, r *http.Request) {
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
	
	goal, err := db.GetSavingsGoalByID(goalID, userID)
	if err != nil {
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}
	
	// Get user finance info for currency
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load user finance info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Goal": goal,
		"CurrentDate": time.Now().Format("2006-01-02"),
		"UserFinance": userFinance,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/contribute_to_goal_form.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "contribute_to_goal_form", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateSavingsGoalHandler handles updating a savings goal via form
func UpdateSavingsGoalHandler(w http.ResponseWriter, r *http.Request) {
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
	
	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Get existing goal to preserve ID and user
	goal, err := db.GetSavingsGoalByID(goalID, userID)
	if err != nil {
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}
	
	// Update fields from form
	goal.Name = r.FormValue("name")
	goal.Priority = r.FormValue("priority")
	goal.Description = stringPtr(r.FormValue("description"))
	goal.ProductURL = stringPtr(r.FormValue("product_link"))
	goal.ImageURL = stringPtr(r.FormValue("image_url"))
	
	// Parse target amount
	targetAmount, err := strconv.ParseFloat(r.FormValue("target_amount"), 64)
	if err != nil {
		http.Error(w, "Invalid target amount", http.StatusBadRequest)
		return
	}
	goal.TargetAmount = targetAmount
	
	// Current amount is not editable - it's calculated from contributions
	
	// Parse monthly contribution
	monthlyContribStr := r.FormValue("monthly_contribution")
	if monthlyContribStr != "" {
		monthlyContrib, err := strconv.ParseFloat(monthlyContribStr, 64)
		if err == nil {
			goal.MonthlyContribution = monthlyContrib
		}
	}
	
	// Parse target months and calculate target date
	targetMonthsStr := r.FormValue("target_months")
	if targetMonthsStr != "" {
		targetMonths, err := strconv.Atoi(targetMonthsStr)
		if err != nil {
			http.Error(w, "Invalid target months", http.StatusBadRequest)
			return
		}
		if targetMonths > 0 {
			targetDate := time.Now().AddDate(0, targetMonths, 0)
			goal.TargetDate = &targetDate
		}
	}
	
	err = db.UpdateSavingsGoal(goal)
	if err != nil {
		http.Error(w, "Failed to update savings goal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return updated finance dashboard
	FinanceHandler(w, r)
}

// DeleteSavingsGoalHandler deletes a savings goal
func DeleteSavingsGoalHandler(w http.ResponseWriter, r *http.Request) {
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
	
	err = db.DeleteSavingsGoal(goalID, userID)
	if err != nil {
		http.Error(w, "Failed to delete savings goal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return script to refresh safe-to-spend card
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<script>htmx.ajax('GET', '/finance/safe-to-spend-card', {target: '#safe-to-spend-card', swap: 'innerHTML'});</script>`))
}

// SavingsGoalPaymentsHandler shows all payments/contributions for a savings goal
func SavingsGoalPaymentsHandler(w http.ResponseWriter, r *http.Request) {
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
	
	// Get the goal details
	goal, err := db.GetSavingsGoalByID(goalID, userID)
	if err != nil {
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}
	
	// Get all contributions for this goal
	contributions, err := db.GetSavingsGoalContributions(goalID, userID)
	if err != nil {
		http.Error(w, "Failed to load contributions: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get user finance info for currency
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load user finance info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Goal":          goal,
		"Contributions": contributions,
		"CurrentDate":   time.Now().Format("2006-01-02"),
		"UserFinance":   userFinance,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/savings_goal_payments_modal.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "savings_goal_payments_modal", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// EditContributionFormHandler shows the edit form for a contribution
func EditContributionFormHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	vars := mux.Vars(r)
	contributionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid contribution ID", http.StatusBadRequest)
		return
	}
	
	contribution, err := db.GetSavingsContributionByID(contributionID, userID)
	if err != nil {
		http.Error(w, "Contribution not found", http.StatusNotFound)
		return
	}
	
	// Get user finance info for currency
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load user finance info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Convert pointer fields to values for template
	methodValue := ""
	if contribution.Method != nil {
		methodValue = *contribution.Method
	}
	
	notesValue := ""
	if contribution.Notes != nil {
		notesValue = *contribution.Notes
	}
	
	data := map[string]interface{}{
		"Contribution": contribution,
		"CurrentDate":  contribution.ContributionDate.Format("2006-01-02"),
		"MethodValue":  methodValue,
		"NotesValue":   notesValue,
		"UserFinance":  userFinance,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/edit_contribution_form.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "edit_contribution_form", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateContributionHandler handles updating a contribution
func UpdateContributionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	vars := mux.Vars(r)
	contributionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid contribution ID", http.StatusBadRequest)
		return
	}
	
	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	
	contributionDate, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}
	
	method := r.FormValue("method")
	if method == "" {
		method = "manual"
	}
	
	notes := r.FormValue("notes")
	
	// Get the goal ID to redirect back to payments modal
	contribution, err := db.GetSavingsContributionByID(contributionID, userID)
	if err != nil {
		http.Error(w, "Contribution not found", http.StatusNotFound)
		return
	}
	
	err = db.UpdateSavingsContribution(contributionID, userID, amount, contributionDate, method, notes)
	if err != nil {
		http.Error(w, "Failed to update contribution: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Redirect back to payments modal
	http.Redirect(w, r, "/finance/goals/"+contribution.SavingsGoalID.String()+"/payments", http.StatusSeeOther)
}

// DeleteContributionHandler deletes a contribution
func DeleteContributionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	vars := mux.Vars(r)
	contributionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid contribution ID", http.StatusBadRequest)
		return
	}
	
	// Get the contribution first to know which goal it belongs to
	contribution, err := db.GetSavingsContributionByID(contributionID, userID)
	if err != nil {
		http.Error(w, "Contribution not found", http.StatusNotFound)
		return
	}
	
	err = db.DeleteSavingsContribution(contributionID, userID)
	if err != nil {
		http.Error(w, "Failed to delete contribution: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Now call the SavingsGoalPaymentsHandler to return the refreshed modal
	// Create a new request to simulate the payments modal request
	goalID := contribution.SavingsGoalID.String()
	
	// Set the goal ID in the request context for the payments handler
	routeVars := map[string]string{"id": goalID}
	r = mux.SetURLVars(r, routeVars)
	
	// Call the payments handler directly to return the updated modal
	SavingsGoalPaymentsHandler(w, r)
}

// SafeToSpendCardHandler returns just the safe-to-spend card content
func SafeToSpendCardHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get the full dashboard data (we need the same calculations)
	dashboard, err := db.GetFinanceDashboard(userID)
	if err != nil {
		http.Error(w, "Failed to load finance data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/safe_to_spend_card.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "safe_to_spend_card", dashboard)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ArchiveSavingsGoalHandler handles archiving a savings goal
func ArchiveSavingsGoalHandler(w http.ResponseWriter, r *http.Request) {
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
	
	err = db.ArchiveSavingsGoal(goalID, userID)
	if err != nil {
		http.Error(w, "Failed to archive goal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return script to refresh safe-to-spend card
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<script>htmx.ajax('GET', '/finance/safe-to-spend-card', {target: '#safe-to-spend-card', swap: 'innerHTML'});</script>`))
}

// ArchivedSavingsGoalsHandler shows all archived savings goals
func ArchivedSavingsGoalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get archived goals
	archivedGoals, err := db.GetArchivedSavingsGoals(userID)
	if err != nil {
		http.Error(w, "Failed to load archived goals: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get user finance info for currency
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load user finance info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Add currency to each archived goal
	for i := range archivedGoals {
		archivedGoals[i].Currency = userFinance.Currency
	}
	
	data := map[string]interface{}{
		"ArchivedGoals":   archivedGoals,
		"UserFinance":     userFinance,
		"SuccessMessage":  "", // No success message for regular view
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/archived_goals_modal.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "archived_goals_modal", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// UnarchiveSavingsGoalHandler handles unarchiving a savings goal
func UnarchiveSavingsGoalHandler(w http.ResponseWriter, r *http.Request) {
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
	
	err = db.UnarchiveSavingsGoal(goalID, userID)
	if err != nil {
		http.Error(w, "Failed to unarchive goal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get the goal name for the success message
	goal, err := db.GetSavingsGoalByID(goalID, userID)
	if err != nil {
		// If we can't get the goal name, still refresh the modal
		ArchivedSavingsGoalsHandler(w, r)
		return
	}
	
	// Get archived goals for the refreshed modal
	archivedGoals, err := db.GetArchivedSavingsGoals(userID)
	if err != nil {
		http.Error(w, "Failed to load archived goals: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get user finance info for currency
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load user finance info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Add currency to each archived goal
	for i := range archivedGoals {
		archivedGoals[i].Currency = userFinance.Currency
	}
	
	data := map[string]interface{}{
		"ArchivedGoals": archivedGoals,
		"UserFinance":   userFinance,
		"SuccessMessage": "✓ " + goal.Name + " has been restored to active goals!",
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/archived_goals_modal.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "archived_goals_modal", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// SavingsGoalsContainerHandler returns just the savings goals container content
func SavingsGoalsContainerHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get savings goals (this already excludes archived goals)
	goals, err := db.GetSavingsGoals(userID)
	if err != nil {
		http.Error(w, "Failed to load savings goals: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get user finance info for currency
	userFinance, err := db.GetUserFinance(userID)
	if err != nil {
		http.Error(w, "Failed to load user finance info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Add currency to each savings goal
	for i := range goals {
		goals[i].Currency = userFinance.Currency
	}
	
	data := map[string]interface{}{
		"ActiveSavingsGoals": goals,
		"UserFinance":        userFinance,
	}
	
	tmpl := template.Must(template.ParseFiles(
		"templates/partials/savings_goals_container.html",
		"templates/partials/savings_goal_item.html",
	))
	
	err = tmpl.ExecuteTemplate(w, "savings_goals_container", data)
	if err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}