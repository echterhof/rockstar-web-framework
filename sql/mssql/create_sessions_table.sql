-- Create sessions table for MSSQL (SQL Server)
-- Stores user session data with expiration tracking
-- Uses NVARCHAR for Unicode string fields and NVARCHAR(MAX) for JSON data
-- MSSQL uses DATETIME2 for better precision than DATETIME

IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sessions]') AND type in (N'U'))
BEGIN
    CREATE TABLE sessions (
        id NVARCHAR(255) PRIMARY KEY,
        user_id NVARCHAR(255) NOT NULL,
        tenant_id NVARCHAR(255) NOT NULL,
        data NVARCHAR(MAX),
        expires_at DATETIME2 NOT NULL,
        created_at DATETIME2 DEFAULT GETDATE(),
        updated_at DATETIME2 DEFAULT GETDATE(),
        ip_address NVARCHAR(45),
        user_agent NVARCHAR(MAX)
    );
END;
