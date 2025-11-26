-- Delete an access token from the database (PostgreSQL)
-- Parameters: $1=token (the token value to delete)
-- PostgreSQL uses $1, $2, etc. for parameter placeholders

DELETE FROM access_tokens
WHERE token = $1;
