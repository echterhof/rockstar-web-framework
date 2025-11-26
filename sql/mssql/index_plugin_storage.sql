-- Create indexes for plugin_storage table (MSSQL)
-- Improves query performance for storage lookups

-- Index on plugin_name for efficient plugin storage lookups
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_storage_plugin' AND object_id = OBJECT_ID('plugin_storage'))
BEGIN
    CREATE INDEX idx_plugin_storage_plugin ON plugin_storage(plugin_name);
END;

-- Index on storage_key for efficient key lookups
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_storage_key' AND object_id = OBJECT_ID('plugin_storage'))
BEGIN
    CREATE INDEX idx_plugin_storage_key ON plugin_storage(storage_key);
END;
