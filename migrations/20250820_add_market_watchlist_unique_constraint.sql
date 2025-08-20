-- Migration: Add unique constraint to market_watchlist table
-- Date: 2025-08-20
-- Description: Add unique constraint on (user_id, symbol) to prevent duplicate entries in watchlist

-- Add unique constraint to prevent duplicate user-symbol combinations
ALTER TABLE market_watchlist ADD CONSTRAINT market_watchlist_user_symbol_unique UNIQUE (user_id, symbol);

-- Note: This constraint is required for the ON CONFLICT clause in the AddToWatchlist function
-- It ensures each user can only have one entry per symbol in their watchlist