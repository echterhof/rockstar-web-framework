-- Load a tenant by hostname from the database
-- Parameters: hostname (as JSON string, e.g., '"example.com"')
-- Returns: id, name, hosts (JSON), config (JSON), is_active, created_at, updated_at, max_users, max_storage, max_requests
-- Uses json_each to search within the hosts JSON array
-- Note: json_each.value returns the raw string value, so we compare directly without json()

SELECT id, name, hosts, config, is_active, created_at, updated_at, max_users, max_storage, max_requests
FROM tenants
WHERE is_active = 1
  AND EXISTS (
    SELECT 1
    FROM json_each(tenants.hosts)
    WHERE json_each.value = ?
  );
