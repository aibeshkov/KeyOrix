-- Migration: Add secret sharing schema
-- Version: 005

-- Add owner_id and is_shared columns to secret_nodes table
ALTER TABLE secret_nodes ADD COLUMN owner_id INTEGER REFERENCES users(id);
ALTER TABLE secret_nodes ADD COLUMN is_shared BOOLEAN DEFAULT FALSE;

-- Create index on owner_id for better query performance
CREATE INDEX idx_secret_nodes_owner_id ON secret_nodes(owner_id);

-- Create share_records table
CREATE TABLE share_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    secret_id INTEGER NOT NULL,
    owner_id INTEGER NOT NULL,
    recipient_id INTEGER NOT NULL,
    is_group BOOLEAN DEFAULT FALSE,
    permission TEXT DEFAULT 'read',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    FOREIGN KEY (secret_id) REFERENCES secret_nodes(id),
    FOREIGN KEY (owner_id) REFERENCES users(id),
    FOREIGN KEY (recipient_id) REFERENCES users(id)
);

-- Create indexes for better query performance
CREATE INDEX idx_share_records_secret_id ON share_records(secret_id);
CREATE INDEX idx_share_records_owner_id ON share_records(owner_id);
CREATE INDEX idx_share_records_recipient_id ON share_records(recipient_id);
CREATE INDEX idx_share_records_deleted_at ON share_records(deleted_at);

-- Update existing secrets to set owner_id to the creator
UPDATE secret_nodes SET owner_id = (
    SELECT id FROM users WHERE username = 'admin' LIMIT 1
) WHERE owner_id IS NULL;

-- Add a trigger to automatically set is_shared flag when shares are created
CREATE TRIGGER update_secret_shared_status_insert
AFTER INSERT ON share_records
BEGIN
    UPDATE secret_nodes SET is_shared = TRUE WHERE id = NEW.secret_id;
END;

-- Add a trigger to update is_shared flag when all shares are deleted
CREATE TRIGGER update_secret_shared_status_delete
AFTER DELETE ON share_records
BEGIN
    UPDATE secret_nodes SET is_shared = (
        SELECT EXISTS(SELECT 1 FROM share_records WHERE secret_id = OLD.secret_id AND deleted_at IS NULL)
    ) WHERE id = OLD.secret_id;
END;