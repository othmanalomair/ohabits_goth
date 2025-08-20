-- Migration: Reset market_watchlist display_order to sequential values
-- Date: 2025-08-20
-- Description: Reset display_order to 1, 2, 3, 4... for each user's watchlist

-- Reset display_order to sequential values starting from 1 for each user
WITH ordered_watchlist AS (
    SELECT 
        id,
        user_id,
        ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at) as new_order
    FROM market_watchlist
)
UPDATE market_watchlist 
SET display_order = ordered_watchlist.new_order
FROM ordered_watchlist 
WHERE market_watchlist.id = ordered_watchlist.id;

-- Note: This resets all users' watchlist display orders to start from 1
-- and increment sequentially based on when items were added (created_at)