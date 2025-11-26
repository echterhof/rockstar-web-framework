-- Create plugin_hooks table for MSSQL (SQL Server)
-- Stores plugin hook registrations and execution statistics
-- Tracks hook type, priority, and performance metrics
-- Foreign key references plugins table

IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[plugin_hooks]') AND type in (N'U'))
BEGIN
    CREATE TABLE plugin_hooks (
        id BIGINT IDENTITY(1,1) PRIMARY KEY,
        plugin_name NVARCHAR(255) NOT NULL,
        hook_type NVARCHAR(50) NOT NULL,
        priority INT DEFAULT 0,
        execution_count INT DEFAULT 0,
        total_duration_ms INT DEFAULT 0,
        error_count INT DEFAULT 0,
        FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
    );
END;
