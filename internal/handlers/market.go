package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
	"ohabits.com/internal/db"
)

// GetMarketDataHandler returns the market data as HTML partial
func GetMarketDataHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		log.Printf("Error getting user ID from context")
		w.WriteHeader(http.StatusUnauthorized)
		renderMarketError(w, "Unauthorized")
		return
	}

	// Get market data for user's watchlist
	marketData, err := db.GetWatchlistWithMarketData(r.Context(), db.DB, userID)
	if err != nil {
		log.Printf("Error getting market data: %v", err)
		renderMarketError(w, "Failed to load market data")
		return
	}

	// If user has no visible items, check if they have any watchlist entries at all
	if len(marketData) == 0 {
		// Check if user has any watchlist entries (visible or not)
		watchlist, err := db.GetUserWatchlist(r.Context(), db.DB, userID)
		if err != nil {
			log.Printf("Error checking user watchlist: %v", err)
			renderMarketError(w, "Failed to check watchlist")
			return
		}

		// Only initialize if user has no watchlist entries at all
		if len(watchlist) == 0 {
			log.Printf("User %v has no market watchlist, initializing defaults", userID)
			err = db.InitializeDefaultWatchlist(r.Context(), db.DB, userID)
			if err != nil {
				log.Printf("Error initializing default watchlist: %v", err)
				renderMarketError(w, "Failed to initialize watchlist")
				return
			}

			// Try to get data again
			marketData, err = db.GetWatchlistWithMarketData(r.Context(), db.DB, userID)
			if err != nil {
				log.Printf("Error getting market data after initialization: %v", err)
				renderMarketError(w, "Failed to load market data")
				return
			}
		}
		// If user has entries but none are visible, show empty state (no initialization)
	}

	// Render the crypto prices partial
	tmpl, err := template.ParseFiles("templates/partials/crypto_prices.html")
	if err != nil {
		log.Printf("Error parsing crypto prices template: %v", err)
		renderMarketError(w, "Template error")
		return
	}

	err = tmpl.Execute(w, marketData)
	if err != nil {
		log.Printf("Error executing crypto prices template: %v", err)
		renderMarketError(w, "Template execution error")
		return
	}
}

// renderMarketError renders an error message for the market data section
func renderMarketError(w http.ResponseWriter, message string) {
	html := `<div class="crypto-error">` + message + `</div>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}