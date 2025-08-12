package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

// NotificationCountHandler returns the count of unread notifications
func NotificationCountHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	count, err := db.GetUnreadNotificationCount(db.DB, userID)
	if err != nil {
		http.Error(w, "Failed to get notification count", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	if count > 0 {
		if count > 99 {
			w.Write([]byte("99+"))
		} else {
			w.Write([]byte(strconv.Itoa(count)))
		}
	}
}

// NotificationListHandler returns the list of notifications
func NotificationListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get query parameters for pagination
	limit := 20 // Default limit
	offset := 0 // Default offset
	
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}
	
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	
	notifications, err := db.GetNotifications(db.DB, userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	
	if len(notifications) == 0 {
		w.Write([]byte(`
			<div class="notification-empty">
				<img src="/static/images/svg/bell.svg" alt="No notifications" />
				<p>No notifications yet</p>
			</div>
		`))
		return
	}
	
	for _, notification := range notifications {
		timeAgo := formatTimeAgo(notification.CreatedAt)
		unreadClass := ""
		if !notification.Read {
			unreadClass = " unread"
		}
		
		iconClass := "notification-icon"
		iconSvg := "bell.svg"
		
		// Set icon based on notification type
		if notification.Type == "new_episodes" {
			iconClass += " episode"
			iconSvg = "tv.svg"
		}
		
		notificationHTML := `
			<div class="notification-item` + unreadClass + `" id="notification-` + notification.ID.String() + `">
				<div class="` + iconClass + `">
					<img src="/static/images/svg/` + iconSvg + `" alt="` + notification.Type + `" />
				</div>
				<div class="notification-content">
					<div class="notification-title">` + notification.Title + `</div>
					<div class="notification-message">` + notification.Message + `</div>
					<div class="notification-time">` + timeAgo + `</div>
				</div>
				<button class="notification-dismiss" onclick="dismissNotification('` + notification.ID.String() + `', event)" title="Dismiss">
					<img src="/static/images/svg/x-dark.svg" alt="Dismiss" />
				</button>
			</div>
		`
		w.Write([]byte(notificationHTML))
	}
}

// DismissNotificationHandler marks a notification as read (keeps for log)
func DismissNotificationHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	notificationIDStr := vars["id"]
	
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}
	
	// Mark as read (keep for log, don't delete)
	err = db.MarkNotificationAsRead(db.DB, notificationID, userID)
	if err != nil {
		http.Error(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	// Return empty content to remove the notification item from the DOM
}

// NotificationAPIHandler provides JSON API endpoints
func NotificationAPIHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	switch r.Method {
	case "GET":
		// Return notifications as JSON
		limit := 20
		offset := 0
		
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
				limit = parsed
			}
		}
		
		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}
		
		notifications, err := db.GetNotifications(db.DB, userID, limit, offset)
		if err != nil {
			http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"notifications": notifications,
			"count":         len(notifications),
		})
		
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// formatTimeAgo converts a time to a human-readable "time ago" format
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return strconv.Itoa(minutes) + " minutes ago"
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return strconv.Itoa(hours) + " hours ago"
	}
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return strconv.Itoa(days) + " days ago"
	}
	if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return strconv.Itoa(weeks) + " weeks ago"
	}
	
	// For older than a month, show the actual date
	return t.Format("Jan 2, 2006")
}