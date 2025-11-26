-- Clear all storage entries for a specific plugin
-- Parameters: plugin_name
-- Removes all key-value pairs for the given plugin

DELETE FROM plugin_storage
WHERE plugin_name = @p1;
