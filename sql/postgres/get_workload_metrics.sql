-- Retrieve workload metrics for a tenant within a time range (PostgreSQL)
-- Parameters: tenant_id, from_timestamp, to_timestamp
-- Returns metrics ordered by timestamp descending (most recent first)

SELECT 
    id, timestamp, tenant_id, user_id, request_id, duration_ms, context_size, 
    memory_usage, cpu_usage, path, method, status_code, response_size, error_message
FROM workload_metrics
WHERE tenant_id = $1 AND timestamp BETWEEN $2 AND $3
ORDER BY timestamp DESC;
