-- Create indexes for sessions table (PostgreSQL)
-- Improves query performance for common access patterns

-- Index on expires_at for efficient cleanup of expired sessions
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Index on user_id for efficient user session lookups
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);

-- Index on tenant_id for efficient tenant session lookups
CREATE INDEX IF NOT EXISTS idx_sessions_tenant ON sessions(tenant_id);
