-- Migration for episode synchronization system
-- This tracks when shows were last updated to avoid API spam

-- Add last_episode_sync and last_info_sync columns to shows table
ALTER TABLE shows ADD COLUMN IF NOT EXISTS last_episode_sync TIMESTAMP;
ALTER TABLE shows ADD COLUMN IF NOT EXISTS last_info_sync TIMESTAMP;

-- Create sync_settings table to store global sync configuration
CREATE TABLE IF NOT EXISTS sync_settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(50) UNIQUE NOT NULL,
    setting_value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert default sync settings
INSERT INTO sync_settings (setting_key, setting_value) VALUES 
    ('episode_sync_interval_hours', '6'),
    ('info_sync_interval_hours', '24'),
    ('max_concurrent_syncs', '3'),
    ('last_global_sync', '2000-01-01 00:00:00')
ON CONFLICT (setting_key) DO NOTHING;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_shows_last_episode_sync ON shows(last_episode_sync);
CREATE INDEX IF NOT EXISTS idx_shows_last_info_sync ON shows(last_info_sync);
CREATE INDEX IF NOT EXISTS idx_shows_status ON shows(status);

-- Create sync_logs table to track sync operations
CREATE TABLE IF NOT EXISTS sync_logs (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    show_id UUID REFERENCES shows(id) ON DELETE CASCADE,
    sync_type VARCHAR(20) NOT NULL, -- 'episodes' or 'info'
    status VARCHAR(20) NOT NULL, -- 'success', 'error', 'skipped'
    episodes_added INTEGER DEFAULT 0,
    info_updated BOOLEAN DEFAULT FALSE,
    error_message TEXT,
    sync_duration_ms INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sync_logs_created_at ON sync_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_sync_logs_show_id ON sync_logs(show_id);