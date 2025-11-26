-- Save or update a value in plugin storage
-- Parameters: plugin_name, storage_key, storage_value, updated_at
-- Uses INSERT OR REPLACE for SQLite upsert semantics

INSERT INTO plugin_storage (plugin_name, storage_key, storage_value, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT (plugin_name, storage_key)
DO UPDATE SET
    storage_value = excluded.storage_value,
    updated_at = excluded.updated_at;
