-- Projects and Tasks Tables
-- Add projects management system to the database

-- Projects table
CREATE TABLE projects (
    id uuid DEFAULT uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT projects_pkey PRIMARY KEY (id),
    CONSTRAINT projects_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Tasks table with hierarchical structure
CREATE TABLE tasks (
    id uuid DEFAULT uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    project_id uuid NOT NULL,
    parent_task_id uuid, -- NULL for top-level tasks, UUID for subtasks
    title text NOT NULL,
    description text,
    status text DEFAULT 'Not Started' NOT NULL,
    priority text DEFAULT 'None' NOT NULL,
    due_date timestamp without time zone,
    completed boolean DEFAULT false NOT NULL,
    display_order integer DEFAULT 0,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT tasks_pkey PRIMARY KEY (id),
    CONSTRAINT tasks_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT tasks_parent_task_id_fkey FOREIGN KEY (parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CONSTRAINT tasks_status_check CHECK (status = ANY (ARRAY['Not Started'::text, 'In Progress'::text, 'Blocked'::text, 'In Review'::text, 'Completed'::text])),
    CONSTRAINT tasks_priority_check CHECK (priority = ANY (ARRAY['None'::text, 'Low'::text, 'Medium'::text, 'High'::text]))
);

-- Task dependencies table (many-to-many relationship)
CREATE TABLE task_dependencies (
    id uuid DEFAULT uuid_generate_v4() NOT NULL,
    task_id uuid NOT NULL,
    depends_on_task_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT task_dependencies_pkey PRIMARY KEY (id),
    CONSTRAINT task_dependencies_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CONSTRAINT task_dependencies_depends_on_task_id_fkey FOREIGN KEY (depends_on_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CONSTRAINT task_dependencies_unique_pair UNIQUE (task_id, depends_on_task_id),
    CONSTRAINT task_dependencies_no_self_reference CHECK (task_id != depends_on_task_id)
);

-- Task comments table
CREATE TABLE task_comments (
    id uuid DEFAULT uuid_generate_v4() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    comment text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT task_comments_pkey PRIMARY KEY (id),
    CONSTRAINT task_comments_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CONSTRAINT task_comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Task attachments table
CREATE TABLE task_attachments (
    id uuid DEFAULT uuid_generate_v4() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    filename text NOT NULL,
    file_path text NOT NULL,
    file_size bigint NOT NULL,
    mime_type text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT task_attachments_pkey PRIMARY KEY (id),
    CONSTRAINT task_attachments_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CONSTRAINT task_attachments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_projects_user_id ON projects USING btree (user_id);
CREATE INDEX idx_tasks_user_id ON tasks USING btree (user_id);
CREATE INDEX idx_tasks_project_id ON tasks USING btree (project_id);
CREATE INDEX idx_tasks_parent_task_id ON tasks USING btree (parent_task_id);
CREATE INDEX idx_tasks_status ON tasks USING btree (status);
CREATE INDEX idx_tasks_priority ON tasks USING btree (priority);
CREATE INDEX idx_tasks_due_date ON tasks USING btree (due_date);
CREATE INDEX idx_tasks_completed ON tasks USING btree (completed);
CREATE INDEX idx_task_dependencies_task_id ON task_dependencies USING btree (task_id);
CREATE INDEX idx_task_dependencies_depends_on_task_id ON task_dependencies USING btree (depends_on_task_id);
CREATE INDEX idx_task_comments_task_id ON task_comments USING btree (task_id);
CREATE INDEX idx_task_attachments_task_id ON task_attachments USING btree (task_id);