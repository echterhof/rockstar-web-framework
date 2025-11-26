-- Create indexes for sessions table (MSSQL)
-- Improves query performance for common access patterns

-- Index on expires_at for efficient cleanup of expired sessions
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_sessions_expires' AND object_id = OBJECT_ID('sessions'))
BEGIN
    CREATE INDEX idx_sessions_expires ON sessions(expires_at);
END;

-- Index on user_id for efficient user session lookups
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_sessions_user' AND object_id = OBJECT_ID('sessions'))
BEGIN
    CREATE INDEX idx_sessions_user ON sessions(user_id);
END;

-- Index on tenant_id for efficient tenant session lookups
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_sessions_tenant' AND object_id = OBJECT_ID('sessions'))
BEGIN
    CREATE INDEX idx_sessions_tenant ON sessions(tenant_id);
END;
