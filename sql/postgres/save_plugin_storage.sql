-- Save or update a value in plugin storage
-- Parameters: plugin_name, storage_key, storage_value, updated_at
-- Uses INSERT ... ON CONFLICT for PostgreSQL upsert semantics

INSERT INTO plugin_storage (plugin_name, storage_key, storage_value, updated_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (plugin_name, storage_key)
DO UPDATE SET
    storage_value = EXCLUDED.storage_value,
    updated_at = EXCLUDED.updated_at;
