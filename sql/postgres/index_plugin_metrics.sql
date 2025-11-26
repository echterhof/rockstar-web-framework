-- Create indexes for plugin_metrics table (PostgreSQL)
-- Improves query performance for metrics lookups and time-based queries

-- Index on plugin_name for efficient plugin metrics lookups
CREATE INDEX IF NOT EXISTS idx_plugin_metrics_plugin ON plugin_metrics(plugin_name);

-- Index on metric_name for filtering metrics by name
CREATE INDEX IF NOT EXISTS idx_plugin_metrics_name ON plugin_metrics(metric_name);

-- Index on recorded_at for time-based queries and sorting
CREATE INDEX IF NOT EXISTS idx_plugin_metrics_recorded ON plugin_metrics(recorded_at);
