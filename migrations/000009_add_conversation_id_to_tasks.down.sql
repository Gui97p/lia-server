DROP INDEX tasks_user_trigger_created_idx;

ALTER TABLE tasks
DROP COLUMN conversation_id;
