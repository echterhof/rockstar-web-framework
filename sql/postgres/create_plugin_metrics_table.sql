-- Create plugin_metrics table for PostgreSQL
-- Stores plugin performance and custom metrics
-- Tracks metric values over time with optional tags
-- Foreign key references plugins table

CREATE TABLE IF NOT EXISTS plugin_metrics (
    id BIGSERIAL PRIMARY KEY,
    plugin_name VARCHAR(255) NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    tags JSONB,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
);
