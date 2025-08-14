-- Add collapsed field to tasks table for expand/collapse functionality
ALTER TABLE tasks ADD COLUMN collapsed boolean DEFAULT false NOT NULL;

-- Add index for better performance on task ordering
CREATE INDEX IF NOT EXISTS idx_tasks_project_display_order ON tasks(project_id, display_order);
CREATE INDEX IF NOT EXISTS idx_tasks_parent_display_order ON tasks(parent_task_id, display_order) WHERE parent_task_id IS NOT NULL;