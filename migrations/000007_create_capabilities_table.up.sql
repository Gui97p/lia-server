CREATE TABLE capabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    parameters JSONB NOT NULL,
    trust_level TEXT NOT NULL CHECK (trust_level IN ('anonymous', 'identified', 'authenticated', 'trusted')),
    source TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER capabilities_updated_at
BEFORE UPDATE ON capabilities
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
