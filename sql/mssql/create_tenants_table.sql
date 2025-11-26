-- Create tenants table for multi-tenant architecture
-- Stores tenant information including hosts, configuration, and resource limits
-- Uses NVARCHAR(MAX) for JSON storage (MSSQL doesn't have a dedicated JSON type)

IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'tenants')
BEGIN
    CREATE TABLE tenants (
        id NVARCHAR(255) PRIMARY KEY,
        name NVARCHAR(255) NOT NULL,
        hosts NVARCHAR(MAX),  -- JSON array of hostnames stored as string
        config NVARCHAR(MAX),  -- JSON object for tenant configuration stored as string
        is_active BIT DEFAULT 1,  -- BIT type for BOOLEAN (1=true, 0=false)
        created_at DATETIME2 DEFAULT GETDATE(),
        updated_at DATETIME2 DEFAULT GETDATE(),
        max_users INT DEFAULT 0,
        max_storage BIGINT DEFAULT 0,
        max_requests BIGINT DEFAULT 0
    );
END;
