-- Delete a session from the database (MSSQL)
-- Parameters: @p1 = id (session ID)
-- Removes the session regardless of expiration status

DELETE FROM sessions WHERE id = ?;
