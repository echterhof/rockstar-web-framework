-- Create indexes for plugin_hooks table (SQLite)
-- Improves query performance for hook lookups and filtering

-- Index on plugin_name for efficient plugin hook lookups
CREATE INDEX IF NOT EXISTS idx_plugin_hooks_plugin ON plugin_hooks(plugin_name);

-- Index on hook_type for filtering hooks by type
CREATE INDEX IF NOT EXISTS idx_plugin_hooks_type ON plugin_hooks(hook_type);

-- Index on priority for ordering hooks by priority
CREATE INDEX IF NOT EXISTS idx_plugin_hooks_priority ON plugin_hooks(priority);
