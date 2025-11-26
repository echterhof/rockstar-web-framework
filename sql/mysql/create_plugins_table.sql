-- Create plugins table for MySQL
-- Stores plugin metadata and configuration
-- Uses BIGINT AUTO_INCREMENT for auto-incrementing IDs
-- MySQL supports JSON type for structured data

CREATE TABLE IF NOT EXISTS plugins (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    author VARCHAR(255),
    enabled BOOLEAN DEFAULT TRUE,
    loaded_at TIMESTAMP NULL,
    config JSON,
    permissions JSON,
    status VARCHAR(50),
    error_count INT DEFAULT 0,
    last_error TEXT,
    last_error_at TIMESTAMP NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
