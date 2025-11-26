-- Create indexes for plugin_events table (MSSQL)
-- Improves query performance for event lookups and filtering

-- Index on event_name for efficient event lookups
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_events_name' AND object_id = OBJECT_ID('plugin_events'))
BEGIN
    CREATE INDEX idx_plugin_events_name ON plugin_events(event_name);
END;

-- Index on publisher_plugin for finding events by publisher
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_events_publisher' AND object_id = OBJECT_ID('plugin_events'))
BEGIN
    CREATE INDEX idx_plugin_events_publisher ON plugin_events(publisher_plugin);
END;

-- Index on subscriber_plugin for finding events by subscriber
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_plugin_events_subscriber' AND object_id = OBJECT_ID('plugin_events'))
BEGIN
    CREATE INDEX idx_plugin_events_subscriber ON plugin_events(subscriber_plugin);
END;
