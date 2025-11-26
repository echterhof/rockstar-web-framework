-- Create rate_limits table for PostgreSQL
-- Stores rate limiting counters with timestamps for sliding window rate limiting
-- Uses BIGSERIAL for auto-incrementing ID
-- PostgreSQL uses TIMESTAMP for timestamp storage

CREATE TABLE IF NOT EXISTS rate_limits (
    id BIGSERIAL PRIMARY KEY,
    rate_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
