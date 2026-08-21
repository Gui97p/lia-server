ALTER TABLE users
ADD COLUMN groq_api_key_encrypted TEXT;

UPDATE users
SET groq_api_key_encrypted = providers.encrypted_key
FROM providers
WHERE users.id = providers.user_id AND providers.encrypted_key IS NOT NULL AND providers.provider = 'groq';

DROP TRIGGER providers_updated_at ON providers;
DROP TABLE providers;
