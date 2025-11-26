-- Create plugin_storage table for PostgreSQL
-- Stores plugin-specific key-value data
-- Each plugin has isolated storage namespace
-- Foreign key references plugins table

CREATE TABLE IF NOT EXISTS plugin_storage (
    id BIGSERIAL PRIMARY KEY,
    plugin_name VARCHAR(255) NOT NULL,
    storage_key VARCHAR(255) NOT NULL,
    storage_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (plugin_name, storage_key),
    FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
);
