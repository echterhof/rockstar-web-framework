-- Save or update a session in the database (SQLite)
-- Parameters: id, user_id, tenant_id, data (JSON), expires_at, created_at, updated_at, ip_address, user_agent
-- Uses INSERT OR REPLACE for SQLite upsert semantics
-- Note: INSERT OR REPLACE will delete and reinsert, which resets created_at
-- For preserving created_at, use INSERT ... ON CONFLICT ... DO UPDATE

INSERT INTO sessions (
    id, user_id, tenant_id, data, expires_at, created_at, updated_at, ip_address, user_agent
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT(id) DO UPDATE SET
    user_id = excluded.user_id,
    tenant_id = excluded.tenant_id,
    data = excluded.data,
    expires_at = excluded.expires_at,
    updated_at = excluded.updated_at,
    ip_address = excluded.ip_address,
    user_agent = excluded.user_agent;
