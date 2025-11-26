-- Create sessions table for SQLite
-- Stores user session data with expiration tracking
-- Uses TEXT for VARCHAR compatibility and INTEGER for BIGINT
-- SQLite uses AUTOINCREMENT for auto-incrementing primary keys

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    data TEXT,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT
);
