-- Remove expired access tokens from the database (PostgreSQL)
-- Parameters: $1=current_time (timestamp to compare against expires_at)
-- Deletes all tokens where expires_at is less than or equal to the current time
-- PostgreSQL uses $1, $2, etc. for parameter placeholders

DELETE FROM access_tokens
WHERE expires_at <= $1;
