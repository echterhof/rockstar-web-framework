-- Create workload_metrics table for PostgreSQL
-- Stores performance and usage metrics for monitoring and analysis
-- Uses BIGSERIAL for auto-incrementing primary key
-- Uses DOUBLE PRECISION for floating point numbers

CREATE TABLE IF NOT EXISTS workload_metrics (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    request_id VARCHAR(255),
    duration_ms BIGINT,
    context_size BIGINT,
    memory_usage BIGINT,
    cpu_usage DOUBLE PRECISION,
    path VARCHAR(500),
    method VARCHAR(10),
    status_code INTEGER,
    response_size BIGINT,
    error_message TEXT
);
