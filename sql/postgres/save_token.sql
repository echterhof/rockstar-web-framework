-- Save or update an access token in the database (PostgreSQL)
-- Parameters: $1=token, $2=user_id, $3=tenant_id, $4=scopes (JSON), $5=expires_at, $6=created_at
-- Uses INSERT ... ON CONFLICT ... DO UPDATE for PostgreSQL upsert semantics
-- EXCLUDED refers to the values that would have been inserted
-- PostgreSQL uses $1, $2, etc. for parameter placeholders

INSERT INTO access_tokens (
    token, user_id, tenant_id, scopes, expires_at, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT(token) DO UPDATE SET
    user_id = EXCLUDED.user_id,
    tenant_id = EXCLUDED.tenant_id,
    scopes = EXCLUDED.scopes,
    expires_at = EXCLUDED.expires_at;
