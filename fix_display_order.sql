
-- SQL script to normalize display orders to be relative to parent levels
UPDATE tasks SET display_order = (
    SELECT ROW_NUMBER() OVER (
        ORDER BY 
            CASE 
                WHEN tasks.display_order = 0 THEN tasks.created_at 
                ELSE NULL 
            END,
            tasks.display_order,
            tasks.created_at
    )
    FROM tasks t2 
    WHERE t2.project_id = tasks.project_id 
      AND t2.user_id = tasks.user_id
      AND (
          (t2.parent_task_id IS NULL AND tasks.parent_task_id IS NULL) OR 
          (t2.parent_task_id = tasks.parent_task_id)
      )
      AND t2.id <= tasks.id
)
WHERE project_id = '82489704-8779-4e9b-a047-fba521d6f27d';

