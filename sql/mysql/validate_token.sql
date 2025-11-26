-- Validate an access token and return it if not expired (MySQL)
-- Parameters: token (the token value to validate), current_time (timestamp to check expiration)
-- Returns: token, user_id, tenant_id, scopes, expires_at, created_at
-- Only returns tokens that have not yet expired

SELECT token, user_id, tenant_id, scopes, expires_at, created_at
FROM access_tokens
WHERE token = ? AND expires_at > ?;
