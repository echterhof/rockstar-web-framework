-- Increment rate limit counter for a key (PostgreSQL)
-- Parameters: rate_key, created_at
-- Inserts a new rate limit entry to track an API call or request

INSERT INTO rate_limits (rate_key, created_at) VALUES ($1, $2);
