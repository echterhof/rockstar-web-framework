-- Remove expired access tokens from the database (MSSQL)
-- Parameters: @p1=current_time (timestamp to compare against expires_at)
-- Deletes all tokens where expires_at is less than or equal to the current time

DELETE FROM access_tokens
WHERE expires_at <= ?;
