-- Create plugin_storage table for MySQL
-- Stores plugin-specific key-value data
-- Each plugin has isolated storage namespace
-- Foreign key references plugins table

CREATE TABLE IF NOT EXISTS plugin_storage (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plugin_name VARCHAR(255) NOT NULL,
    storage_key VARCHAR(255) NOT NULL,
    storage_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_plugin_key (plugin_name, storage_key),
    FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
