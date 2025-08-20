-- Migration: Update market_watchlist type constraint to include new asset types
-- Date: 2025-08-20
-- Description: Add support for 'commodity' and 'index' types in addition to existing types

-- Drop the existing check constraint
ALTER TABLE market_watchlist DROP CONSTRAINT market_watchlist_type_check;

-- Add updated check constraint with new asset types
ALTER TABLE market_watchlist ADD CONSTRAINT market_watchlist_type_check 
    CHECK (type = ANY (ARRAY['stock'::text, 'crypto'::text, 'forex'::text, 'commodity'::text, 'index'::text]));

-- Note: This allows the following asset types:
-- - crypto: Bitcoin, Ethereum, XRP, Dogecoin
-- - commodity: Gold, Oil
-- - index: S&P 500
-- - stock: Boursa Kuwait, individual stocks
-- - forex: Currency pairs (future expansion)