-- Save or update a session in the database (MySQL)
-- Parameters: id, user_id, tenant_id, data (JSON), expires_at, created_at, updated_at, ip_address, user_agent,
--            data (for update), expires_at (for update), updated_at (for update), ip_address (for update), user_agent (for update)
-- Uses INSERT ... ON DUPLICATE KEY UPDATE for MySQL upsert semantics

INSERT INTO sessions (
    id, user_id, tenant_id, data, expires_at, created_at, updated_at, ip_address, user_agent
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON DUPLICATE KEY UPDATE
    data = VALUES(data),
    expires_at = VALUES(expires_at),
    updated_at = VALUES(updated_at),
    ip_address = VALUES(ip_address),
    user_agent = VALUES(user_agent);
