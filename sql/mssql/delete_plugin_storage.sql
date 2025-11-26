-- Delete a specific key-value pair from plugin storage
-- Parameters: plugin_name, storage_key
-- Removes the entry for the given plugin and key

DELETE FROM plugin_storage
WHERE plugin_name = @p1 AND storage_key = @p2;
