-- Migration: Add markdown_notes table
-- Date: 2025-08-20
-- Description: Create table for large markdown notes with titles, auto-save, and search functionality

-- Create markdown_notes table
CREATE TABLE markdown_notes (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- Add indexes for better performance
CREATE INDEX idx_markdown_notes_user_id ON markdown_notes(user_id);
CREATE INDEX idx_markdown_notes_title ON markdown_notes(title);
CREATE INDEX idx_markdown_notes_created_at ON markdown_notes(created_at DESC);
CREATE INDEX idx_markdown_notes_updated_at ON markdown_notes(updated_at DESC);

-- Create full-text search index for content and title
CREATE INDEX idx_markdown_notes_search ON markdown_notes USING gin(to_tsvector('english', title || ' ' || content));

-- Note: This creates a new notes system separate from the existing daily notes
-- The existing 'notes' table remains for daily journal entries