-- Create indexes for plugin_metrics table (MSSQL)
-- Improves query performance for metrics lookups and time-based queries

-- Index on plugin_name for efficient plugin metrics lookups
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_metrics_plugin' AND object_id = OBJECT_ID('plugin_metrics'))
BEGIN
    CREATE INDEX idx_plugin_metrics_plugin ON plugin_metrics(plugin_name);
END;

-- Index on metric_name for filtering metrics by name
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_metrics_name' AND object_id = OBJECT_ID('plugin_metrics'))
BEGIN
    CREATE INDEX idx_plugin_metrics_name ON plugin_metrics(metric_name);
END;

-- Index on recorded_at for time-based queries and sorting
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_metrics_recorded' AND object_id = OBJECT_ID('plugin_metrics'))
BEGIN
    CREATE INDEX idx_plugin_metrics_recorded ON plugin_metrics(recorded_at);
END;
