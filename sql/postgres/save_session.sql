-- Save or update a session in the database (PostgreSQL)
-- Parameters: id, user_id, tenant_id, data (JSON), expires_at, created_at, updated_at, ip_address, user_agent
-- Uses INSERT ... ON CONFLICT ... DO UPDATE for PostgreSQL upsert semantics
-- EXCLUDED refers to the values that would have been inserted

INSERT INTO sessions (
    id, user_id, tenant_id, data, expires_at, created_at, updated_at, ip_address, user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT(id) DO UPDATE SET
    user_id = EXCLUDED.user_id,
    tenant_id = EXCLUDED.tenant_id,
    data = EXCLUDED.data,
    expires_at = EXCLUDED.expires_at,
    updated_at = EXCLUDED.updated_at,
    ip_address = EXCLUDED.ip_address,
    user_agent = EXCLUDED.user_agent;
