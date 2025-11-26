-- Create access_tokens table for MSSQL (SQL Server)
-- Stores API access tokens with expiration tracking and scope management
-- Uses NVARCHAR for Unicode string fields and NVARCHAR(MAX) for JSON storage
-- MSSQL uses DATETIME2 for datetime fields (more precise than DATETIME)

CREATE TABLE access_tokens (
    token NVARCHAR(255) PRIMARY KEY,
    user_id NVARCHAR(255) NOT NULL,
    tenant_id NVARCHAR(255) NOT NULL,
    scopes NVARCHAR(MAX),
    expires_at DATETIME2 NOT NULL,
    created_at DATETIME2 DEFAULT GETDATE()
);
