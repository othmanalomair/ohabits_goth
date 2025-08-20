package api

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
	"ohabits.com/internal/db"
	"ohabits.com/internal/services/market"
)

// GetMarketData retrieves market data for user's watchlist
func GetMarketData(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's market data
	marketData, err := db.GetWatchlistWithMarketData(r.Context(), db.DB, userID)
	if err != nil {
		log.Printf("Error getting watchlist with market data: %v", err)
		http.Error(w, "Failed to get market data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(marketData)
}

// ToggleWatchlistItem adds or removes a symbol from user's watchlist
func ToggleWatchlistItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data from HTMX request
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	request := struct {
		Symbol string
		Name   string
		Type   string
	}{
		Symbol: r.FormValue("symbol"),
		Name:   r.FormValue("name"),
		Type:   r.FormValue("type"),
	}

	if request.Symbol == "" || request.Name == "" || request.Type == "" {
		http.Error(w, "Symbol, name, and type are required", http.StatusBadRequest)
		return
	}

	_, err := db.ToggleWatchlistItem(r.Context(), db.DB, userID, request.Symbol, request.Name, request.Type)
	if err != nil {
		log.Printf("Error toggling watchlist item: %v", err)
		http.Error(w, "Failed to toggle watchlist item", http.StatusInternalServerError)
		return
	}

	// Return updated crypto settings HTML
	cryptoSettings, err := getUserCryptoSettings(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting crypto settings after toggle: %v", err)
		renderCryptoError(w, "Failed to load updated crypto settings")
		return
	}

	// Render the crypto settings partial
	tmpl, err := template.ParseFiles("templates/partials/crypto_settings.html")
	if err != nil {
		log.Printf("Error parsing crypto settings template: %v", err)
		renderCryptoError(w, "Template error")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, cryptoSettings)
	if err != nil {
		log.Printf("Error executing crypto settings template: %v", err)
		renderCryptoError(w, "Template execution error")
		return
	}
}

// GetUserWatchlist retrieves user's complete watchlist as HTML for the settings page
func GetUserWatchlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all available cryptos with watchlist status
	cryptoSettings, err := getUserCryptoSettings(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting crypto settings: %v", err)
		renderCryptoError(w, "Failed to load crypto settings")
		return
	}

	// Render the crypto settings partial
	tmpl, err := template.ParseFiles("templates/partials/crypto_settings.html")
	if err != nil {
		log.Printf("Error parsing crypto settings template: %v", err)
		renderCryptoError(w, "Template error")
		return
	}

	err = tmpl.Execute(w, cryptoSettings)
	if err != nil {
		log.Printf("Error executing crypto settings template: %v", err)
		renderCryptoError(w, "Template execution error")
		return
	}
}

// CryptoSetting represents an asset (crypto/stock/commodity) with its watchlist status
type CryptoSetting struct {
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	InWatchlist  bool   `json:"in_watchlist"`
	DisplayOrder int    `json:"display_order"`
}

// getUserCryptoSettings gets all available assets with their watchlist status
func getUserCryptoSettings(ctx context.Context, userID uuid.UUID) ([]CryptoSetting, error) {
	// Define available assets (crypto + traditional markets)
	availableAssets := map[string]struct {
		name      string
		assetType string
	}{
		// Cryptocurrencies
		"BTC":    {"Bitcoin", "crypto"},
		"ETH":    {"Ethereum", "crypto"},
		"XRP":    {"XRP", "crypto"},
		"DOGE":   {"Dogecoin", "crypto"},
		// Traditional Markets
		"GOLD":   {"Gold ETF", "commodity"},
		"OIL":    {"Crude Oil", "commodity"},
		"SPX":    {"S&P 500", "index"},
		"BOURSA": {"Boursa Kuwait", "index"},
	}

	// Get user's current watchlist
	watchlist, err := db.GetUserWatchlist(ctx, db.DB, userID)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup of watchlist items with display order
	watchlistMap := make(map[string]struct {
		inWatchlist  bool
		displayOrder int
	})
	for _, item := range watchlist {
		watchlistMap[item.Symbol] = struct {
			inWatchlist  bool
			displayOrder int
		}{true, item.DisplayOrder}
	}

	// Build the settings list - watchlist items first (ordered), then disabled items
	var settings []CryptoSetting
	
	// First add watchlist items in display order
	for _, item := range watchlist {
		if asset, exists := availableAssets[item.Symbol]; exists {
			settings = append(settings, CryptoSetting{
				Symbol:       item.Symbol,
				Name:         asset.name,
				Type:         asset.assetType,
				InWatchlist:  true,
				DisplayOrder: item.DisplayOrder,
			})
		}
	}
	
	// Then add non-watchlist items
	for symbol, asset := range availableAssets {
		if !watchlistMap[symbol].inWatchlist {
			settings = append(settings, CryptoSetting{
				Symbol:       symbol,
				Name:         asset.name,
				Type:         asset.assetType,
				InWatchlist:  false,
				DisplayOrder: 0,
			})
		}
	}

	return settings, nil
}

// renderCryptoError renders an error message for the crypto settings section
func renderCryptoError(w http.ResponseWriter, message string) {
	html := `<div class="crypto-error">` + message + `</div>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// RefreshMarketData manually triggers a market data refresh
func RefreshMarketData(w http.ResponseWriter, r *http.Request) {
	// This endpoint will be used to manually refresh market data
	client := market.NewCoinCapClient()
	symbols := client.GetSupportedSymbols()

	marketData, err := client.GetAssets(symbols)
	if err != nil {
		log.Printf("Error fetching market data: %v", err)
		http.Error(w, "Failed to fetch market data", http.StatusInternalServerError)
		return
	}

	// Update database with new market data
	for _, data := range marketData {
		dbData := db.MarketData{
			Symbol:        data.Symbol,
			CurrentPrice:  data.CurrentPrice,
			ChangeAmount:  data.ChangeAmount,
			ChangePercent: data.ChangePercent,
			Volume:        data.Volume,
			MarketCap:     data.MarketCap,
			LastUpdated:   data.LastUpdated,
		}

		if err := db.UpdateMarketData(r.Context(), db.DB, dbData); err != nil {
			log.Printf("Error updating market data for %s: %v", data.Symbol, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Market data refreshed successfully",
		"count":   len(marketData),
	})
}

// MoveWatchlistItemUp moves a watchlist item up in the display order
func MoveWatchlistItemUp(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data from HTMX request
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	symbol := r.FormValue("symbol")
	if symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	err := db.MoveWatchlistItemUp(r.Context(), db.DB, userID, symbol)
	if err != nil {
		log.Printf("Error moving watchlist item up: %v", err)
		http.Error(w, "Failed to move item up", http.StatusInternalServerError)
		return
	}

	// Return updated crypto settings HTML
	cryptoSettings, err := getUserCryptoSettings(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting crypto settings after move: %v", err)
		renderCryptoError(w, "Failed to load updated crypto settings")
		return
	}

	// Render the crypto settings partial
	tmpl, err := template.ParseFiles("templates/partials/crypto_settings.html")
	if err != nil {
		log.Printf("Error parsing crypto settings template: %v", err)
		renderCryptoError(w, "Template error")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, cryptoSettings)
	if err != nil {
		log.Printf("Error executing crypto settings template: %v", err)
		renderCryptoError(w, "Template execution error")
		return
	}
}

// MoveWatchlistItemDown moves a watchlist item down in the display order
func MoveWatchlistItemDown(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data from HTMX request
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	symbol := r.FormValue("symbol")
	if symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	err := db.MoveWatchlistItemDown(r.Context(), db.DB, userID, symbol)
	if err != nil {
		log.Printf("Error moving watchlist item down: %v", err)
		http.Error(w, "Failed to move item down", http.StatusInternalServerError)
		return
	}

	// Return updated crypto settings HTML
	cryptoSettings, err := getUserCryptoSettings(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting crypto settings after move: %v", err)
		renderCryptoError(w, "Failed to load updated crypto settings")
		return
	}

	// Render the crypto settings partial
	tmpl, err := template.ParseFiles("templates/partials/crypto_settings.html")
	if err != nil {
		log.Printf("Error parsing crypto settings template: %v", err)
		renderCryptoError(w, "Template error")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, cryptoSettings)
	if err != nil {
		log.Printf("Error executing crypto settings template: %v", err)
		renderCryptoError(w, "Template execution error")
		return
	}
}
