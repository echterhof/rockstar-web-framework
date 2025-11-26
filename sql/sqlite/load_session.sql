-- Load a session from the database (SQLite)
-- Parameters: id (session ID), current_timestamp (to check expiration)
-- Returns: id, user_id, tenant_id, data, expires_at, created_at, updated_at, ip_address, user_agent
-- Only returns sessions that have not expired

SELECT 
    id, 
    user_id, 
    tenant_id, 
    data, 
    expires_at, 
    created_at, 
    updated_at, 
    ip_address, 
    user_agent
FROM sessions 
WHERE id = ? AND expires_at > ?;
