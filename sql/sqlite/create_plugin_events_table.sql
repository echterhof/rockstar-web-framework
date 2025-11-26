-- Create plugin_events table for SQLite
-- Stores plugin event subscriptions and publications
-- Tracks which plugins publish and subscribe to events
-- Foreign keys reference plugins table

CREATE TABLE IF NOT EXISTS plugin_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_name TEXT NOT NULL,
    publisher_plugin TEXT NOT NULL,
    subscriber_plugin TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (publisher_plugin) REFERENCES plugins(name) ON DELETE CASCADE,
    FOREIGN KEY (subscriber_plugin) REFERENCES plugins(name) ON DELETE CASCADE
);
