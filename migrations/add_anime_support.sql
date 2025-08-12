-- Migration to add anime support to shows table
-- This adds a show_type field to distinguish between TV shows and anime
-- and renames tvmaze_id to external_id to be more generic

BEGIN;

-- Add show_type column to distinguish between TV shows and anime
ALTER TABLE public.shows ADD COLUMN show_type varchar(20) DEFAULT 'tv' NOT NULL;

-- Add constraint to ensure show_type is either 'tv' or 'anime'
ALTER TABLE public.shows ADD CONSTRAINT shows_type_check CHECK (show_type IN ('tv', 'anime'));

-- Rename tvmaze_id to external_id to be more generic (supports both TVMaze ID and MAL ID)
ALTER TABLE public.shows RENAME COLUMN tvmaze_id TO external_id;

-- Update the index name to reflect the column rename
DROP INDEX IF EXISTS idx_shows_tvmaze_id;
CREATE INDEX idx_shows_external_id ON public.shows USING btree (external_id);

-- Add composite index for efficient lookups by external_id and show_type
CREATE INDEX idx_shows_external_id_type ON public.shows USING btree (external_id, show_type);

-- Also update the episodes table to remove the dependency on TVMaze specifically
-- Rename tvmaze_id to external_id in episodes table as well
ALTER TABLE public.episodes RENAME COLUMN tvmaze_id TO external_id;

-- Update the index name for episodes
DROP INDEX IF EXISTS idx_episodes_tvmaze_id;
CREATE INDEX idx_episodes_external_id ON public.episodes USING btree (external_id);

-- Add show_type to episodes table for consistency (though it can be derived from shows)
ALTER TABLE public.episodes ADD COLUMN show_type varchar(20) DEFAULT 'tv' NOT NULL;
ALTER TABLE public.episodes ADD CONSTRAINT episodes_type_check CHECK (show_type IN ('tv', 'anime'));

COMMIT;