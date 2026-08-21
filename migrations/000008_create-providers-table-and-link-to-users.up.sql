CREATE TABLE providers (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    groq_api_key_encrypted TEXT,
    gemini_api_key_encrypted TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER providers_updated_at
BEFORE UPDATE ON providers
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO providers (user_id, groq_api_key_encrypted)
SELECT id, groq_api_key_encrypted FROM users
WHERE groq_api_key_encrypted IS NOT NULL;

ALTER TABLE users
DROP COLUMN groq_api_key_encrypted;
