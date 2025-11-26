-- Delete an access token from the database (MSSQL)
-- Parameters: @p1=token (the token value to delete)

DELETE FROM access_tokens
WHERE token = ?;
