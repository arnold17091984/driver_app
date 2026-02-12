-- Remove phone_number column and revert role constraint

DROP INDEX IF EXISTS idx_users_phone_number;
ALTER TABLE users DROP COLUMN IF EXISTS phone_number;

-- Delete any passenger users first
DELETE FROM users WHERE role = 'passenger';

-- Revert role constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin','dispatcher','viewer','driver'));
