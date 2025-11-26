-- Create indexes for plugin_events table (SQLite)
-- Improves query performance for event lookups and filtering

-- Index on event_name for efficient event lookups
CREATE INDEX IF NOT EXISTS idx_plugin_events_name ON plugin_events(event_name);

-- Index on publisher_plugin for finding events by publisher
CREATE INDEX IF NOT EXISTS idx_plugin_events_publisher ON plugin_events(publisher_plugin);

-- Index on subscriber_plugin for finding events by subscriber
CREATE INDEX IF NOT EXISTS idx_plugin_events_subscriber ON plugin_events(subscriber_plugin);
