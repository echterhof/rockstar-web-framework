-- Save or update an access token in the database (MySQL)
-- Parameters: token, user_id, tenant_id, scopes (JSON), expires_at, created_at
-- Uses INSERT ... ON DUPLICATE KEY UPDATE for MySQL upsert semantics
-- VALUES() function references the values from the INSERT clause

INSERT INTO access_tokens (
    token, user_id, tenant_id, scopes, expires_at, created_at
) VALUES (
    ?, ?, ?, ?, ?, ?
)
ON DUPLICATE KEY UPDATE
    user_id = VALUES(user_id),
    tenant_id = VALUES(tenant_id),
    scopes = VALUES(scopes),
    expires_at = VALUES(expires_at);
