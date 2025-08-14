-- Add file attachment support to tasks table
ALTER TABLE tasks ADD COLUMN attachment_filename TEXT;
ALTER TABLE tasks ADD COLUMN attachment_file_path TEXT;
ALTER TABLE tasks ADD COLUMN attachment_content_type TEXT;
ALTER TABLE tasks ADD COLUMN attachment_file_size BIGINT;

-- Create an index for better performance when querying tasks with attachments
CREATE INDEX idx_tasks_with_attachments ON tasks(id) WHERE attachment_filename IS NOT NULL;