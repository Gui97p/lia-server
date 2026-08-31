ALTER TABLE tasks
ADD COLUMN conversation_id UUID NOT NULL DEFAULT gen_random_uuid();

CREATE INDEX tasks_user_trigger_created_idx ON tasks (user_id, trigger_type, created_at DESC);
