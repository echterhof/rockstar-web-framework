-- Clean up expired sessions from the database (SQLite)
-- Parameters: current_timestamp
-- Removes all sessions where expires_at is less than or equal to the current time

DELETE FROM sessions WHERE expires_at <= ?;
