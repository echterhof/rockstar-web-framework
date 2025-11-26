-- Create indexes for rate_limits table (SQLite)
-- Improves query performance for rate limiting checks

-- Composite index on rate_key and created_at for efficient rate limit checks
-- This index supports both the WHERE clause filtering and time-based queries
CREATE INDEX IF NOT EXISTS idx_rate_limits_key_time ON rate_limits(rate_key, created_at);
