-- Save or update a value in plugin storage
-- Parameters: plugin_name, storage_key, storage_value, updated_at
-- Uses INSERT ... ON DUPLICATE KEY UPDATE for MySQL upsert semantics

INSERT INTO plugin_storage (plugin_name, storage_key, storage_value, updated_at)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    storage_value = VALUES(storage_value),
    updated_at = VALUES(updated_at);
