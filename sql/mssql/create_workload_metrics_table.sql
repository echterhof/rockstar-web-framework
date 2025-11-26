-- Create workload_metrics table for MSSQL
-- Stores performance and usage metrics for monitoring and analysis
-- Uses BIGINT IDENTITY for auto-incrementing primary key
-- Uses FLOAT for floating point numbers

CREATE TABLE workload_metrics (
    id BIGINT IDENTITY(1,1) PRIMARY KEY,
    timestamp DATETIME2 DEFAULT GETDATE(),
    tenant_id NVARCHAR(255) NOT NULL,
    user_id NVARCHAR(255),
    request_id NVARCHAR(255),
    duration_ms BIGINT,
    context_size BIGINT,
    memory_usage BIGINT,
    cpu_usage FLOAT,
    path NVARCHAR(500),
    method NVARCHAR(10),
    status_code INT,
    response_size BIGINT,
    error_message NVARCHAR(MAX)
);
