-- Migration to add filler and recap fields to episodes table

BEGIN;

-- Add filler column for anime episodes
ALTER TABLE public.episodes ADD COLUMN filler boolean DEFAULT false NOT NULL;

-- Add recap column for anime episodes  
ALTER TABLE public.episodes ADD COLUMN recap boolean DEFAULT false NOT NULL;

-- Add indexes for efficient filtering
CREATE INDEX idx_episodes_filler ON public.episodes USING btree (filler);
CREATE INDEX idx_episodes_recap ON public.episodes USING btree (recap);

-- Add composite index for filler/recap filtering
CREATE INDEX idx_episodes_filler_recap ON public.episodes USING btree (filler, recap);

COMMIT;