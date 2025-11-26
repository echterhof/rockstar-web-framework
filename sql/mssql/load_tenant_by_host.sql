-- Load a tenant by hostname from the database
-- Parameters: hostname (as JSON string, e.g., '"example.com"')
-- Returns: id, name, hosts (JSON), config (JSON), is_active, created_at, updated_at, max_users, max_storage, max_requests
-- Uses OPENJSON to search within the hosts JSON array

SELECT id, name, hosts, config, is_active, created_at, updated_at, max_users, max_storage, max_requests
FROM tenants
WHERE is_active = 1
  AND EXISTS (
    SELECT 1
    FROM OPENJSON(hosts)
    WHERE value = JSON_QUERY(?)
  );
