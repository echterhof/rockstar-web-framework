-- Create access_tokens table for SQLite
-- Stores API access tokens with expiration tracking and scope management
-- Uses TEXT for VARCHAR compatibility
-- SQLite uses TEXT for JSON storage

CREATE TABLE IF NOT EXISTS access_tokens (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    scopes TEXT,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
