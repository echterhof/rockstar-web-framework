-- Create plugins table for PostgreSQL
-- Stores plugin metadata and configuration
-- Uses BIGSERIAL for auto-incrementing IDs
-- PostgreSQL supports JSONB type for better performance

CREATE TABLE IF NOT EXISTS plugins (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    author VARCHAR(255),
    enabled BOOLEAN DEFAULT TRUE,
    loaded_at TIMESTAMP,
    config JSONB,
    permissions JSONB,
    status VARCHAR(50),
    error_count INTEGER DEFAULT 0,
    last_error TEXT,
    last_error_at TIMESTAMP
);
