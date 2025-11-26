-- Create indexes for plugin_hooks table (MSSQL)
-- Improves query performance for hook lookups and filtering

-- Index on plugin_name for efficient plugin hook lookups
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_hooks_plugin' AND object_id = OBJECT_ID('plugin_hooks'))
BEGIN
    CREATE INDEX idx_plugin_hooks_plugin ON plugin_hooks(plugin_name);
END;

-- Index on hook_type for filtering hooks by type
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_hooks_type' AND object_id = OBJECT_ID('plugin_hooks'))
BEGIN
    CREATE INDEX idx_plugin_hooks_type ON plugin_hooks(hook_type);
END;

-- Index on priority for ordering hooks by priority
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_hooks_priority' AND object_id = OBJECT_ID('plugin_hooks'))
BEGIN
    CREATE INDEX idx_plugin_hooks_priority ON plugin_hooks(priority);
END;
