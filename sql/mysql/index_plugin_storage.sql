-- Create indexes for plugin_storage table (MySQL)
-- Improves query performance for storage lookups

-- Index on plugin_name for efficient plugin storage lookups
CREATE INDEX idx_plugin_storage_plugin ON plugin_storage(plugin_name);

-- Index on storage_key for efficient key lookups
CREATE INDEX idx_plugin_storage_key ON plugin_storage(storage_key);
