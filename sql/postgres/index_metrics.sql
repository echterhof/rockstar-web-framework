-- Create indexes for workload_metrics table (PostgreSQL)
-- Improves query performance for common access patterns

-- Index on timestamp for efficient time-range queries
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON workload_metrics(timestamp);

-- Index on tenant_id for efficient tenant-specific queries
CREATE INDEX IF NOT EXISTS idx_metrics_tenant ON workload_metrics(tenant_id);

-- Index on user_id for efficient user-specific queries
CREATE INDEX IF NOT EXISTS idx_metrics_user ON workload_metrics(user_id);

-- Composite index on tenant_id and timestamp for optimized range queries per tenant
CREATE INDEX IF NOT EXISTS idx_metrics_tenant_timestamp ON workload_metrics(tenant_id, timestamp);
