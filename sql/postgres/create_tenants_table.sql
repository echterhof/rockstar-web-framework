-- Create tenants table for multi-tenant architecture
-- Stores tenant information including hosts, configuration, and resource limits
-- Uses JSONB type for hosts array and config object (more efficient than JSON in PostgreSQL)

CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    hosts JSONB,  -- JSONB array of hostnames (more efficient than JSON)
    config JSONB,  -- JSONB object for tenant configuration
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    max_users INTEGER DEFAULT 0,
    max_storage BIGINT DEFAULT 0,
    max_requests BIGINT DEFAULT 0
);
