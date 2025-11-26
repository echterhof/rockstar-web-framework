-- Retrieve workload metrics for a tenant within a time range (SQLite)
-- Parameters: tenant_id, from_timestamp, to_timestamp
-- Returns metrics ordered by timestamp descending (most recent first)

SELECT 
    id, timestamp, tenant_id, user_id, request_id, duration_ms, context_size, 
    memory_usage, cpu_usage, path, method, status_code, response_size, error_message
FROM workload_metrics
WHERE tenant_id = ? AND timestamp BETWEEN ? AND ?
ORDER BY timestamp DESC;
