package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UpdateMarketData updates or inserts market data for a symbol
func UpdateMarketData(ctx context.Context, db *pgxpool.Pool, marketData MarketData) error {
	query := `
		INSERT INTO market_data (symbol, current_price, change_amount, change_percent, volume, market_cap, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (symbol)
		DO UPDATE SET
			current_price = EXCLUDED.current_price,
			change_amount = EXCLUDED.change_amount,
			change_percent = EXCLUDED.change_percent,
			volume = EXCLUDED.volume,
			market_cap = EXCLUDED.market_cap,
			last_updated = EXCLUDED.last_updated
	`
	
	_, err := db.Exec(ctx, query,
		marketData.Symbol,
		marketData.CurrentPrice,
		marketData.ChangeAmount,
		marketData.ChangePercent,
		marketData.Volume,
		marketData.MarketCap,
		marketData.LastUpdated,
	)
	
	return err
}

// GetMarketData retrieves market data for specific symbols
func GetMarketData(ctx context.Context, db *pgxpool.Pool, symbols []string) ([]MarketData, error) {
	if len(symbols) == 0 {
		return []MarketData{}, nil
	}

	query := `
		SELECT id, symbol, current_price, change_amount, change_percent, volume, market_cap, last_updated
		FROM market_data
		WHERE symbol = ANY($1)
		ORDER BY symbol
	`
	
	rows, err := db.Query(ctx, query, symbols)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var marketData []MarketData
	for rows.Next() {
		var data MarketData
		err := rows.Scan(
			&data.ID,
			&data.Symbol,
			&data.CurrentPrice,
			&data.ChangeAmount,
			&data.ChangePercent,
			&data.Volume,
			&data.MarketCap,
			&data.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		marketData = append(marketData, data)
	}

	return marketData, rows.Err()
}

// GetUserWatchlist retrieves user's market watchlist
func GetUserWatchlist(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) ([]MarketWatchlist, error) {
	query := `
		SELECT id, user_id, symbol, name, type, display_order, created_at
		FROM market_watchlist
		WHERE user_id = $1
		ORDER BY display_order, created_at
	`
	
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchlist []MarketWatchlist
	for rows.Next() {
		var item MarketWatchlist
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Symbol,
			&item.Name,
			&item.Type,
			&item.DisplayOrder,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		watchlist = append(watchlist, item)
	}

	return watchlist, rows.Err()
}

// AddToWatchlist adds a symbol to user's watchlist
func AddToWatchlist(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, symbol, name, symbolType string) error {
	// Get the next display order
	var maxOrder int
	err := db.QueryRow(ctx, 
		"SELECT COALESCE(MAX(display_order), 0) FROM market_watchlist WHERE user_id = $1", 
		userID).Scan(&maxOrder)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO market_watchlist (user_id, symbol, name, type, display_order)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, symbol) DO NOTHING
	`
	
	_, err = db.Exec(ctx, query, userID, symbol, name, symbolType, maxOrder+1)
	return err
}

// RemoveFromWatchlist removes a symbol from user's watchlist
func RemoveFromWatchlist(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, symbol string) error {
	query := `DELETE FROM market_watchlist WHERE user_id = $1 AND symbol = $2`
	_, err := db.Exec(ctx, query, userID, symbol)
	return err
}

// ToggleWatchlistItem toggles a symbol in user's watchlist (add if not exists, remove if exists)
func ToggleWatchlistItem(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, symbol, name, symbolType string) (bool, error) {
	// Check if item exists
	var exists bool
	err := db.QueryRow(ctx, 
		"SELECT EXISTS(SELECT 1 FROM market_watchlist WHERE user_id = $1 AND symbol = $2)", 
		userID, symbol).Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists {
		// Remove from watchlist
		err = RemoveFromWatchlist(ctx, db, userID, symbol)
		return false, err // false means removed
	} else {
		// Add to watchlist
		err = AddToWatchlist(ctx, db, userID, symbol, name, symbolType)
		return true, err // true means added
	}
}

// GetWatchlistWithMarketData retrieves user's watchlist with current market data
func GetWatchlistWithMarketData(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) ([]MarketData, error) {
	query := `
		SELECT 
			md.id, md.symbol, md.current_price, md.change_amount, md.change_percent, 
			md.volume, md.market_cap, md.last_updated
		FROM market_watchlist mw
		JOIN market_data md ON mw.symbol = md.symbol
		WHERE mw.user_id = $1
		ORDER BY mw.display_order, mw.created_at
	`
	
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var marketData []MarketData
	for rows.Next() {
		var data MarketData
		err := rows.Scan(
			&data.ID,
			&data.Symbol,
			&data.CurrentPrice,
			&data.ChangeAmount,
			&data.ChangePercent,
			&data.Volume,
			&data.MarketCap,
			&data.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		marketData = append(marketData, data)
	}

	return marketData, rows.Err()
}

// InitializeDefaultWatchlist sets up default symbols for a new user
func InitializeDefaultWatchlist(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) error {
	defaultAssets := []struct {
		symbol    string
		name      string
		assetType string
	}{
		// Default cryptocurrencies
		{"BTC", "Bitcoin", "crypto"},
		{"ETH", "Ethereum", "crypto"},
		// Default traditional markets
		{"GOLD", "Gold ETF", "commodity"},
		{"SPX", "S&P 500", "index"},
	}

	for _, asset := range defaultAssets {
		err := AddToWatchlist(ctx, db, userID, asset.symbol, asset.name, asset.assetType)
		if err != nil {
			return fmt.Errorf("failed to add %s to watchlist: %w", asset.symbol, err)
		}
	}

	return nil
}

// MoveWatchlistItemUp moves a watchlist item up in the display order
func MoveWatchlistItemUp(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, symbol string) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get the current item's display_order
	var currentOrder int
	err = tx.QueryRow(ctx, 
		"SELECT display_order FROM market_watchlist WHERE user_id = $1 AND symbol = $2",
		userID, symbol).Scan(&currentOrder)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}

	// Find the item with the previous order (to swap with)
	var swapItemID uuid.UUID
	var swapOrder int
	err = tx.QueryRow(ctx,
		`SELECT id, display_order FROM market_watchlist 
		 WHERE user_id = $1 AND display_order < $2 
		 ORDER BY display_order DESC LIMIT 1`,
		userID, currentOrder).Scan(&swapItemID, &swapOrder)
	if err != nil {
		// Already at the top
		return nil
	}

	// Swap the display orders
	_, err = tx.Exec(ctx,
		"UPDATE market_watchlist SET display_order = $1 WHERE user_id = $2 AND symbol = $3",
		swapOrder, userID, symbol)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		"UPDATE market_watchlist SET display_order = $1 WHERE id = $2",
		currentOrder, swapItemID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// MoveWatchlistItemDown moves a watchlist item down in the display order
func MoveWatchlistItemDown(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, symbol string) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get the current item's display_order
	var currentOrder int
	err = tx.QueryRow(ctx, 
		"SELECT display_order FROM market_watchlist WHERE user_id = $1 AND symbol = $2",
		userID, symbol).Scan(&currentOrder)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}

	// Find the item with the next order (to swap with)
	var swapItemID uuid.UUID
	var swapOrder int
	err = tx.QueryRow(ctx,
		`SELECT id, display_order FROM market_watchlist 
		 WHERE user_id = $1 AND display_order > $2 
		 ORDER BY display_order ASC LIMIT 1`,
		userID, currentOrder).Scan(&swapItemID, &swapOrder)
	if err != nil {
		// Already at the bottom
		return nil
	}

	// Swap the display orders
	_, err = tx.Exec(ctx,
		"UPDATE market_watchlist SET display_order = $1 WHERE user_id = $2 AND symbol = $3",
		swapOrder, userID, symbol)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		"UPDATE market_watchlist SET display_order = $1 WHERE id = $2",
		currentOrder, swapItemID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}