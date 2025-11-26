-- Save or update a tenant in the database
-- Parameters: id, name, hosts (JSON), config (JSON), is_active, created_at, updated_at, max_users, max_storage, max_requests
-- Uses INSERT OR REPLACE for SQLite upsert semantics

INSERT OR REPLACE INTO tenants (
    id, name, hosts, config, is_active, created_at, updated_at, max_users, max_storage, max_requests
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);
