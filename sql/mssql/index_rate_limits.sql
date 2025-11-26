-- Create indexes for rate_limits table (MSSQL)
-- Improves query performance for rate limiting checks

-- Composite index on rate_key and created_at for efficient rate limit checks
-- This index supports both the WHERE clause filtering and time-based queries
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_rate_limits_key_time' AND object_id = OBJECT_ID('rate_limits'))
BEGIN
    CREATE INDEX idx_rate_limits_key_time ON rate_limits(rate_key, created_at);
END;
