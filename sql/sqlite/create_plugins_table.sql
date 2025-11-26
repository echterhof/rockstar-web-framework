-- Create plugins table for SQLite
-- Stores plugin metadata and configuration
-- Uses INTEGER PRIMARY KEY AUTOINCREMENT for auto-incrementing IDs
-- SQLite uses TEXT for VARCHAR and JSON storage

CREATE TABLE IF NOT EXISTS plugins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    version TEXT NOT NULL,
    description TEXT,
    author TEXT,
    enabled INTEGER DEFAULT 1,
    loaded_at DATETIME,
    config TEXT,
    permissions TEXT,
    status TEXT,
    error_count INTEGER DEFAULT 0,
    last_error TEXT,
    last_error_at DATETIME
);
