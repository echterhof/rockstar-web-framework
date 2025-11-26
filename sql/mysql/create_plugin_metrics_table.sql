-- Create plugin_metrics table for MySQL
-- Stores plugin performance and custom metrics
-- Tracks metric values over time with optional tags
-- Foreign key references plugins table

CREATE TABLE IF NOT EXISTS plugin_metrics (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plugin_name VARCHAR(255) NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    metric_value DOUBLE NOT NULL,
    tags JSON,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
