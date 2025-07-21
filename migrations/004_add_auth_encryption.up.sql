-- Migration to add encrypted fields for authentication data
-- This migration adds encrypted storage for API secrets, session tokens, and password reset tokens

-- Add encrypted fields to api_clients table
ALTER TABLE api_clients 
ADD COLUMN encrypted_client_secret BLOB,
ADD COLUMN client_secret_metadata JSON;

-- Add encrypted fields to sessions table
ALTER TABLE sessions 
ADD COLUMN encrypted_session_token BLOB,
ADD COLUMN session_token_metadata JSON;

-- Add encrypted fields to api_tokens table
ALTER TABLE api_tokens 
ADD COLUMN encrypted_token BLOB,
ADD COLUMN token_metadata JSON;

-- Add encrypted fields to password_resets table
ALTER TABLE password_resets 
ADD COLUMN encrypted_token BLOB,
ADD COLUMN token_metadata JSON;

-- Create indexes for better performance on encrypted fields
CREATE INDEX idx_api_clients_encrypted_secret ON api_clients(encrypted_client_secret(255));
CREATE INDEX idx_sessions_encrypted_token ON sessions(encrypted_session_token(255));
CREATE INDEX idx_api_tokens_encrypted_token ON api_tokens(encrypted_token(255));
CREATE INDEX idx_password_resets_encrypted_token ON password_resets(encrypted_token(255));

-- Add comments to document the encryption
COMMENT ON COLUMN api_clients.encrypted_client_secret IS 'AES-256-GCM encrypted client secret';
COMMENT ON COLUMN api_clients.client_secret_metadata IS 'Encryption metadata (algorithm, nonce, key version)';
COMMENT ON COLUMN sessions.encrypted_session_token IS 'AES-256-GCM encrypted session token';
COMMENT ON COLUMN sessions.session_token_metadata IS 'Encryption metadata (algorithm, nonce, key version)';
COMMENT ON COLUMN api_tokens.encrypted_token IS 'AES-256-GCM encrypted API token';
COMMENT ON COLUMN api_tokens.token_metadata IS 'Encryption metadata (algorithm, nonce, key version)';
COMMENT ON COLUMN password_resets.encrypted_token IS 'AES-256-GCM encrypted password reset token';
COMMENT ON COLUMN password_resets.token_metadata IS 'Encryption metadata (algorithm, nonce, key version)';