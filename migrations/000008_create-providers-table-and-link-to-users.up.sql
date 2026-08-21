CREATE TABLE providers (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,

    provider TEXT NOT NULL,
    encrypted_key TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, provider)
);

CREATE TRIGGER providers_updated_at
BEFORE UPDATE ON providers
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO providers (user_id, provider, encrypted_key)
SELECT id, 'groq', groq_api_key_encrypted FROM users
WHERE groq_api_key_encrypted IS NOT NULL;

ALTER TABLE users
DROP COLUMN groq_api_key_encrypted;
