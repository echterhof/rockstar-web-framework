-- Load a tenant by ID from the database
-- Parameters: id (tenant ID)
-- Returns: id, name, hosts (JSON), config (JSON), is_active, created_at, updated_at, max_users, max_storage, max_requests

SELECT id, name, hosts, config, is_active, created_at, updated_at, max_users, max_storage, max_requests
FROM tenants
WHERE id = ?;
