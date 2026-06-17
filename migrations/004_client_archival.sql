ALTER TABLE IF EXISTS users
  ADD COLUMN IF NOT EXISTS archived_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_users_archived_at ON users(archived_at);
