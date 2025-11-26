-- Create indexes for plugin_hooks table (MySQL)
-- Improves query performance for hook lookups and filtering

-- Index on plugin_name for efficient plugin hook lookups
CREATE INDEX idx_plugin_hooks_plugin ON plugin_hooks(plugin_name);

-- Index on hook_type for filtering hooks by type
CREATE INDEX idx_plugin_hooks_type ON plugin_hooks(hook_type);

-- Index on priority for ordering hooks by priority
CREATE INDEX idx_plugin_hooks_priority ON plugin_hooks(priority);
