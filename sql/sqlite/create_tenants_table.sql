-- Create tenants table for multi-tenant architecture
-- Stores tenant information including hosts, configuration, and resource limits
-- Uses JSON type for hosts array and config object

CREATE TABLE IF NOT EXISTS tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    hosts TEXT,  -- JSON array of hostnames
    config TEXT,  -- JSON object for tenant configuration
    is_active INTEGER DEFAULT 1,  -- SQLite uses INTEGER for BOOLEAN (1=true, 0=false)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    max_users INTEGER DEFAULT 0,
    max_storage INTEGER DEFAULT 0,
    max_requests INTEGER DEFAULT 0
);
