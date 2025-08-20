-- Migration: Add visible field to market_watchlist
-- Date: 2025-08-20
-- Description: Add visible boolean field to control display without deleting entries

-- Add visible column with default true
ALTER TABLE market_watchlist ADD COLUMN visible BOOLEAN NOT NULL DEFAULT true;

-- Note: This allows users to hide/show market items without losing their display order
-- All existing entries will be visible by default