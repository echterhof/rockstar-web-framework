-- Create workload_metrics table for MySQL
-- Stores performance and usage metrics for monitoring and analysis
-- Uses BIGINT AUTO_INCREMENT for primary key
-- Uses DOUBLE for floating point numbers

CREATE TABLE IF NOT EXISTS workload_metrics (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    request_id VARCHAR(255),
    duration_ms BIGINT,
    context_size BIGINT,
    memory_usage BIGINT,
    cpu_usage DOUBLE,
    path VARCHAR(500),
    method VARCHAR(10),
    status_code INT,
    response_size BIGINT,
    error_message TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
