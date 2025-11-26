-- Create plugin_storage table for SQLite
-- Stores plugin-specific key-value data
-- Each plugin has isolated storage namespace
-- Foreign key references plugins table

CREATE TABLE IF NOT EXISTS plugin_storage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_name TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    storage_value TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (plugin_name, storage_key),
    FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
);
