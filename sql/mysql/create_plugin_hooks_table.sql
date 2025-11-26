-- Create plugin_hooks table for MySQL
-- Stores plugin hook registrations and execution statistics
-- Tracks hook type, priority, and performance metrics
-- Foreign key references plugins table

CREATE TABLE IF NOT EXISTS plugin_hooks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plugin_name VARCHAR(255) NOT NULL,
    hook_type VARCHAR(50) NOT NULL,
    priority INT DEFAULT 0,
    execution_count INT DEFAULT 0,
    total_duration_ms INT DEFAULT 0,
    error_count INT DEFAULT 0,
    FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
