-- Validate an access token and return it if not expired (PostgreSQL)
-- Parameters: $1=token (the token value to validate), $2=current_time (timestamp to check expiration)
-- Returns: token, user_id, tenant_id, scopes, expires_at, created_at
-- Only returns tokens that have not yet expired
-- PostgreSQL uses $1, $2, etc. for parameter placeholders

SELECT token, user_id, tenant_id, scopes, expires_at, created_at
FROM access_tokens
WHERE token = $1 AND expires_at > $2;
