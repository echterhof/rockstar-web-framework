-- Create tenants table for multi-tenant architecture
-- Stores tenant information including hosts, configuration, and resource limits
-- Uses JSON type for hosts array and config object

CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    hosts JSON,  -- JSON array of hostnames
    config JSON,  -- JSON object for tenant configuration
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    max_users INT DEFAULT 0,
    max_storage BIGINT DEFAULT 0,
    max_requests BIGINT DEFAULT 0
);
