-- Add 'passenger' to role constraint and phone_number column for passenger registration

-- Drop existing role constraint and add new one with 'passenger'
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin','dispatcher','viewer','driver','passenger'));

-- Add phone_number column (nullable; only passengers use it)
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number) WHERE phone_number IS NOT NULL;
