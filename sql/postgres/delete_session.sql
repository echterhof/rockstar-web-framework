-- Delete a session from the database (PostgreSQL)
-- Parameters: $1 = id (session ID)
-- Removes the session regardless of expiration status

DELETE FROM sessions WHERE id = $1;
