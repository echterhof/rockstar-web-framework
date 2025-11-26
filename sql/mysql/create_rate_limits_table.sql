-- Create rate_limits table for MySQL
-- Stores rate limiting counters with timestamps for sliding window rate limiting
-- Uses BIGINT AUTO_INCREMENT for auto-incrementing ID
-- MySQL uses TIMESTAMP for timestamp storage with automatic initialization

CREATE TABLE IF NOT EXISTS rate_limits (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    rate_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
