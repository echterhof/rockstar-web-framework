-- Load an access token from the database (PostgreSQL)
-- Parameters: $1=token (the token value to look up)
-- Returns: token, user_id, tenant_id, scopes, expires_at, created_at
-- PostgreSQL uses $1, $2, etc. for parameter placeholders

SELECT token, user_id, tenant_id, scopes, expires_at, created_at
FROM access_tokens
WHERE token = $1;
