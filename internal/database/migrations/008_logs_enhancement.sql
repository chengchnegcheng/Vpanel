-- Migration: 008_logs_enhancement
-- Description: Enhance logs table with request_id, fields, and optimized indexes

-- +migrate Up

-- Add request_id column for request tracing
ALTER TABLE logs ADD COLUMN request_id VARCHAR(100);

-- Add fields column for extra context (JSON encoded)
ALTER TABLE logs ADD COLUMN fields TEXT;

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_logs_level_created ON logs(level, created_at);
CREATE INDEX IF NOT EXISTS idx_logs_source_created ON logs(source, created_at);
CREATE INDEX IF NOT EXISTS idx_logs_user_created ON logs(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_logs_request_id ON logs(request_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_logs_request_id;
DROP INDEX IF EXISTS idx_logs_user_created;
DROP INDEX IF EXISTS idx_logs_source_created;
DROP INDEX IF EXISTS idx_logs_level_created;

-- Note: SQLite doesn't support DROP COLUMN directly
-- For rollback, the columns will remain but indexes will be removed
