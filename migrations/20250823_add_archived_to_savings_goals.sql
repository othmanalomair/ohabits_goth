-- Add archived functionality to savings goals
ALTER TABLE savings_goals ADD COLUMN is_archived BOOLEAN DEFAULT FALSE;

-- Create index for better query performance
CREATE INDEX idx_savings_goals_archived ON savings_goals(user_id, is_archived, is_active);