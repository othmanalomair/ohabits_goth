-- Add user-specific news source preferences
-- This allows each user to choose which news sources they want to see

CREATE TABLE IF NOT EXISTS user_news_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES news_sources(id) ON DELETE CASCADE,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one preference per user per source
    UNIQUE(user_id, source_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_news_preferences_user_id ON user_news_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_user_news_preferences_source_id ON user_news_preferences(source_id);
CREATE INDEX IF NOT EXISTS idx_user_news_preferences_enabled ON user_news_preferences(user_id, is_enabled);

-- Create trigger to update updated_at
CREATE OR REPLACE FUNCTION update_user_news_preferences_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_news_preferences_updated_at
    BEFORE UPDATE ON user_news_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_user_news_preferences_updated_at();