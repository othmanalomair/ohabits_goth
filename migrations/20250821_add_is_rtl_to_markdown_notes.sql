-- Migration: Add is_rtl column to markdown_notes table
-- Date: 2025-08-21
-- Description: Add boolean column to track text direction (RTL/LTR) for markdown notes

-- Add is_rtl column to markdown_notes table
ALTER TABLE markdown_notes 
ADD COLUMN is_rtl BOOLEAN NOT NULL DEFAULT FALSE;

-- Add comment for documentation
COMMENT ON COLUMN markdown_notes.is_rtl IS 'Indicates text direction: true for RTL (Arabic), false for LTR (English)';

-- Note: Default is FALSE (LTR) for existing notes to maintain current behavior