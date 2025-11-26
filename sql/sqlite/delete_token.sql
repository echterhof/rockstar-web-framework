-- Delete an access token from the database (SQLite)
-- Parameters: token (the token value to delete)

DELETE FROM access_tokens
WHERE token = ?;
