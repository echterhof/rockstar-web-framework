-- Save or update a tenant in the database
-- Parameters: id, name, hosts (JSON), config (JSON), is_active, created_at, updated_at, max_users, max_storage, max_requests
-- Uses INSERT ... ON CONFLICT ... DO UPDATE for PostgreSQL upsert semantics

INSERT INTO tenants (
    id, name, hosts, config, is_active, created_at, updated_at, max_users, max_storage, max_requests
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    hosts = EXCLUDED.hosts,
    config = EXCLUDED.config,
    is_active = EXCLUDED.is_active,
    updated_at = EXCLUDED.updated_at,
    max_users = EXCLUDED.max_users,
    max_storage = EXCLUDED.max_storage,
    max_requests = EXCLUDED.max_requests;
