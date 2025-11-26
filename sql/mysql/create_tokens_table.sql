-- Create access_tokens table for MySQL
-- Stores API access tokens with expiration tracking and scope management
-- Uses VARCHAR for string fields and JSON type for structured data
-- MySQL uses TIMESTAMP for datetime fields

CREATE TABLE IF NOT EXISTS access_tokens (
    token VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    scopes JSON,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
