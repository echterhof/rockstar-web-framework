-- List all storage keys for a specific plugin
-- Parameters: plugin_name
-- Returns all keys stored by the plugin

SELECT storage_key
FROM plugin_storage
WHERE plugin_name = ?
ORDER BY storage_key;
