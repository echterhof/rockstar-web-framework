-- Create indexes for plugins table (MSSQL)
-- Improves query performance for common access patterns

-- Index on name for efficient plugin lookups by name
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugins_name' AND object_id = OBJECT_ID('plugins'))
BEGIN
    CREATE INDEX idx_plugins_name ON plugins(name);
END;

-- Index on enabled for filtering active/inactive plugins
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugins_enabled' AND object_id = OBJECT_ID('plugins'))
BEGIN
    CREATE INDEX idx_plugins_enabled ON plugins(enabled);
END;

-- Index on status for filtering plugins by status
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugins_status' AND object_id = OBJECT_ID('plugins'))
BEGIN
    CREATE INDEX idx_plugins_status ON plugins(status);
END;
