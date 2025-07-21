-- Rollback migration for authentication encryption fields

-- Drop indexes
DROP INDEX IF EXISTS idx_password_resets_encrypted_token;
DROP INDEX IF EXISTS idx_api_tokens_encrypted_token;
DROP INDEX IF EXISTS idx_sessions_encrypted_token;
DROP INDEX IF EXISTS idx_api_clients_encrypted_secret;

-- Remove encrypted fields from password_resets table
ALTER TABLE password_resets 
DROP COLUMN IF EXISTS token_metadata,
DROP COLUMN IF EXISTS encrypted_token;

-- Remove encrypted fields from api_tokens table
ALTER TABLE api_tokens 
DROP COLUMN IF EXISTS token_metadata,
DROP COLUMN IF EXISTS encrypted_token;

-- Remove encrypted fields from sessions table
ALTER TABLE sessions 
DROP COLUMN IF EXISTS session_token_metadata,
DROP COLUMN IF EXISTS encrypted_session_token;

-- Remove encrypted fields from api_clients table
ALTER TABLE api_clients 
DROP COLUMN IF EXISTS client_secret_metadata,
DROP COLUMN IF EXISTS encrypted_client_secret;