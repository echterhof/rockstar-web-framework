-- Save or update an access token in the database (SQLite)
-- Parameters: token, user_id, tenant_id, scopes (JSON), expires_at, created_at
-- Uses INSERT ... ON CONFLICT ... DO UPDATE for SQLite upsert semantics
-- Preserves created_at on updates

INSERT INTO access_tokens (
    token, user_id, tenant_id, scopes, expires_at, created_at
) VALUES (
    ?, ?, ?, ?, ?, ?
)
ON CONFLICT(token) DO UPDATE SET
    user_id = excluded.user_id,
    tenant_id = excluded.tenant_id,
    scopes = excluded.scopes,
    expires_at = excluded.expires_at;
