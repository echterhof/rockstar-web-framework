-- Create rate_limits table for SQLite
-- Stores rate limiting counters with timestamps for sliding window rate limiting
-- Uses INTEGER PRIMARY KEY for auto-incrementing ID
-- SQLite uses DATETIME for timestamp storage

CREATE TABLE IF NOT EXISTS rate_limits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rate_key TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
