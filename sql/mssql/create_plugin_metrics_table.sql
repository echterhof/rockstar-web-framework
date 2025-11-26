-- Create plugin_metrics table for MSSQL (SQL Server)
-- Stores plugin performance and custom metrics
-- Tracks metric values over time with optional tags
-- Foreign key references plugins table

IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[plugin_metrics]') AND type in (N'U'))
BEGIN
    CREATE TABLE plugin_metrics (
        id BIGINT IDENTITY(1,1) PRIMARY KEY,
        plugin_name NVARCHAR(255) NOT NULL,
        metric_name NVARCHAR(255) NOT NULL,
        metric_value FLOAT NOT NULL,
        tags NVARCHAR(MAX),
        recorded_at DATETIME2 DEFAULT GETDATE(),
        FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
    );
END;
