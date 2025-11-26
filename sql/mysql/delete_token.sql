-- Delete an access token from the database (MySQL)
-- Parameters: token (the token value to delete)

DELETE FROM access_tokens
WHERE token = ?;
