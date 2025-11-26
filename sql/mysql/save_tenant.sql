-- Save or update a tenant in the database
-- Parameters: id, name, hosts (JSON), config (JSON), is_active, created_at, updated_at, max_users, max_storage, max_requests (x2 for UPDATE clause)
-- Uses INSERT ... ON DUPLICATE KEY UPDATE for MySQL upsert semantics

INSERT INTO tenants (
    id, name, hosts, config, is_active, created_at, updated_at, max_users, max_storage, max_requests
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON DUPLICATE KEY UPDATE
    name = ?,
    hosts = ?,
    config = ?,
    is_active = ?,
    updated_at = ?,
    max_users = ?,
    max_storage = ?,
    max_requests = ?;
