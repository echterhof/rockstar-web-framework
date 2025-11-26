-- Delete a session from the database (MySQL)
-- Parameters: id (session ID)
-- Removes the session regardless of expiration status

DELETE FROM sessions WHERE id = ?;
