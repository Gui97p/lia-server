CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    state TEXT NOT NULL CHECK(state IN ('CREATED', 'PLANNING', 'READY', 'RUNNING', 'WAITING', 'BLOCKED', 'REPLANNING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    workflow JSONB,
    trigger_type TEXT NOT NULL CHECK(trigger_type IN ('user', 'scheduled', 'event')),
    authorized_trust_level TEXT NOT NULL CHECK(authorized_trust_level IN ('anonymous', 'identified', 'authenticated', 'trusted')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER tasks_updated_at
BEFORE UPDATE ON tasks
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX user_tasks_index ON tasks (user_id, created_at);
