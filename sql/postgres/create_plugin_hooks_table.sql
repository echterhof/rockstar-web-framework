-- Create plugin_hooks table for PostgreSQL
-- Stores plugin hook registrations and execution statistics
-- Tracks hook type, priority, and performance metrics
-- Foreign key references plugins table

CREATE TABLE IF NOT EXISTS plugin_hooks (
    id BIGSERIAL PRIMARY KEY,
    plugin_name VARCHAR(255) NOT NULL,
    hook_type VARCHAR(50) NOT NULL,
    priority INTEGER DEFAULT 0,
    execution_count INTEGER DEFAULT 0,
    total_duration_ms INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
);
