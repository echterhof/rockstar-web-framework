-- Load a specific value from plugin storage
-- Parameters: plugin_name, storage_key
-- Returns the storage_value for the given plugin and key

SELECT storage_value
FROM plugin_storage
WHERE plugin_name = ? AND storage_key = ?;
