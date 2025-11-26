-- Load an access token from the database (MySQL)
-- Parameters: token (the token value to look up)
-- Returns: token, user_id, tenant_id, scopes, expires_at, created_at

SELECT token, user_id, tenant_id, scopes, expires_at, created_at
FROM access_tokens
WHERE token = ?;
