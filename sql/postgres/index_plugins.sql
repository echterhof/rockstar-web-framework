-- Create indexes for plugins table (PostgreSQL)
-- Improves query performance for common access patterns

-- Index on name for efficient plugin lookups by name
CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);

-- Index on enabled for filtering active/inactive plugins
CREATE INDEX IF NOT EXISTS idx_plugins_enabled ON plugins(enabled);

-- Index on status for filtering plugins by status
CREATE INDEX IF NOT EXISTS idx_plugins_status ON plugins(status);
