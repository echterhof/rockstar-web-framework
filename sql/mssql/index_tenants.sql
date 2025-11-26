-- Create index on tenants table for active status
-- Improves query performance when filtering by is_active

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_tenants_active' AND object_id = OBJECT_ID('tenants'))
BEGIN
    CREATE INDEX idx_tenants_active ON tenants(is_active);
END;
