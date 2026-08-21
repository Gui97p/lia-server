ALTER TABLE users
ADD COLUMN groq_api_key_encrypted TEXT;

UPDATE users
SET groq_api_key_encrypted = providers.groq_api_key_encrypted
FROM providers
WHERE users.id = providers.user_id AND providers.groq_api_key_encrypted IS NOT NULL;

DROP TRIGGER providers_updated_at ON providers;
DROP TABLE providers;
