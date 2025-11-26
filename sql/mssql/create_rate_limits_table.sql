-- Create rate_limits table for MSSQL
-- Stores rate limiting counters with timestamps for sliding window rate limiting
-- Uses BIGINT IDENTITY for auto-incrementing ID
-- MSSQL uses DATETIME2 for timestamp storage with better precision

IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'rate_limits')
BEGIN
    CREATE TABLE rate_limits (
        id BIGINT IDENTITY(1,1) PRIMARY KEY,
        rate_key NVARCHAR(255) NOT NULL,
        created_at DATETIME2 DEFAULT GETDATE()
    );
END;
