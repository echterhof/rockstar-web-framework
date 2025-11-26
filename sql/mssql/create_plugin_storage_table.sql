-- Create plugin_storage table for MSSQL (SQL Server)
-- Stores plugin-specific key-value data
-- Each plugin has isolated storage namespace
-- Foreign key references plugins table

IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[plugin_storage]') AND type in (N'U'))
BEGIN
    CREATE TABLE plugin_storage (
        id BIGINT IDENTITY(1,1) PRIMARY KEY,
        plugin_name NVARCHAR(255) NOT NULL,
        storage_key NVARCHAR(255) NOT NULL,
        storage_value NVARCHAR(MAX),
        created_at DATETIME2 DEFAULT GETDATE(),
        updated_at DATETIME2 DEFAULT GETDATE(),
        CONSTRAINT unique_plugin_key UNIQUE (plugin_name, storage_key),
        FOREIGN KEY (plugin_name) REFERENCES plugins(name) ON DELETE CASCADE
    );
END;
