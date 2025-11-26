-- Create indexes for access_tokens table (MSSQL)
-- Indexes improve query performance for common access patterns

-- Index on expires_at for efficient cleanup of expired tokens
CREATE INDEX idx_tokens_expires ON access_tokens(expires_at);

-- Index on user_id for looking up all tokens for a user
CREATE INDEX idx_tokens_user ON access_tokens(user_id);

-- Index on tenant_id for looking up all tokens for a tenant
CREATE INDEX idx_tokens_tenant ON access_tokens(tenant_id);
