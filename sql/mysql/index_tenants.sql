-- Create index on tenants table for active status
-- Improves query performance when filtering by is_active

CREATE INDEX idx_tenants_active ON tenants(is_active);
