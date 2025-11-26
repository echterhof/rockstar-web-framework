-- Check rate limit for a given key (PostgreSQL)
-- Parameters: rate_key, window_start (timestamp)
-- Returns the count of rate limit entries within the time window
-- Used to determine if rate limit has been exceeded

SELECT COUNT(*) FROM rate_limits WHERE rate_key = $1 AND created_at > $2;
