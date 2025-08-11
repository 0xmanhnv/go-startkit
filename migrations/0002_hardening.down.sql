-- Revert CHECK constraint and trigger

DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_role_check;


