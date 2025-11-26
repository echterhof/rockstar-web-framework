-- Create workload_metrics table for SQLite
-- Stores performance and usage metrics for monitoring and analysis
-- Uses INTEGER for auto-incrementing primary key (SQLite's AUTOINCREMENT)
-- Uses TEXT for VARCHAR compatibility and REAL for floating point numbers

CREATE TABLE IF NOT EXISTS workload_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    tenant_id TEXT NOT NULL,
    user_id TEXT,
    request_id TEXT,
    duration_ms INTEGER,
    context_size INTEGER,
    memory_usage INTEGER,
    cpu_usage REAL,
    path TEXT,
    method TEXT,
    status_code INTEGER,
    response_size INTEGER,
    error_message TEXT
);
