-- Create plugin_metrics table for SQLite
-- Stores plugin performance and custom metrics
-- Tracks metric values over time with optional tags
-- Foreign key references plugins table

CREATE TABLE IF NOT EXISTS plugin_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_name TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    metric_value REAL NOT NULL,
    tags TEXT,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
);
