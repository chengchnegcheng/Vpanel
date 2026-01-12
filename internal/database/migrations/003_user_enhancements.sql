-- Migration: User model enhancements
-- Version: 003
-- Requirements: 17.6, 17.7, 17.8

-- Add new columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS traffic_limit BIGINT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS traffic_used BIGINT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS force_password_change BOOLEAN DEFAULT 0;

-- Add indexes for new columns
CREATE INDEX IF NOT EXISTS idx_users_expires_at ON users(expires_at);
CREATE INDEX IF NOT EXISTS idx_users_traffic_limit ON users(traffic_limit);
