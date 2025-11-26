-- Create plugins table for MSSQL (SQL Server)
-- Stores plugin metadata and configuration
-- Uses BIGINT IDENTITY for auto-incrementing IDs
-- MSSQL uses NVARCHAR for Unicode strings and NVARCHAR(MAX) for JSON

IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[plugins]') AND type in (N'U'))
BEGIN
    CREATE TABLE plugins (
        id BIGINT IDENTITY(1,1) PRIMARY KEY,
        name NVARCHAR(255) UNIQUE NOT NULL,
        version NVARCHAR(50) NOT NULL,
        description NVARCHAR(MAX),
        author NVARCHAR(255),
        enabled BIT DEFAULT 1,
        loaded_at DATETIME2,
        config NVARCHAR(MAX),
        permissions NVARCHAR(MAX),
        status NVARCHAR(50),
        error_count INT DEFAULT 0,
        last_error NVARCHAR(MAX),
        last_error_at DATETIME2
    );
END;
