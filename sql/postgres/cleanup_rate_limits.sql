-- Clean up old rate limit entries (PostgreSQL)
-- Parameters: cutoff_timestamp
-- Removes rate limit entries older than the specified timestamp
-- Used to maintain sliding window and prevent table growth

DELETE FROM rate_limits WHERE created_at <= $1;
