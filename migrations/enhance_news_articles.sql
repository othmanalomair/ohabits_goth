-- Migration: Enhance news articles with full content and images
-- Date: 2025-08-15

-- Add new columns to news_articles table
ALTER TABLE news_articles 
ADD COLUMN IF NOT EXISTS thumbnail_url TEXT,
ADD COLUMN IF NOT EXISTS full_content TEXT,
ADD COLUMN IF NOT EXISTS author_name TEXT,
ADD COLUMN IF NOT EXISTS keywords TEXT,
ADD COLUMN IF NOT EXISTS date_modified TIMESTAMP WITH TIME ZONE;

-- Create index on author for searching
CREATE INDEX IF NOT EXISTS idx_news_articles_author ON news_articles(author_name);

-- Create index on keywords for searching  
CREATE INDEX IF NOT EXISTS idx_news_articles_keywords ON news_articles USING GIN (to_tsvector('arabic', keywords));

-- Update existing records to set default values
UPDATE news_articles 
SET 
    thumbnail_url = image_url,
    full_content = content,
    date_modified = updated_at
WHERE thumbnail_url IS NULL;

COMMENT ON COLUMN news_articles.thumbnail_url IS 'URL to article thumbnail image';
COMMENT ON COLUMN news_articles.full_content IS 'Complete article body text from JSON-LD';
COMMENT ON COLUMN news_articles.author_name IS 'Article author name from JSON-LD';
COMMENT ON COLUMN news_articles.keywords IS 'Article keywords from JSON-LD';
COMMENT ON COLUMN news_articles.date_modified IS 'Last modification date from JSON-LD';