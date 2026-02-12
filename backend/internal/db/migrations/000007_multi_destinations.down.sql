-- Re-add destination column
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS destination TEXT;

-- Migrate first destination back
UPDATE reservations SET destination = destinations[1] WHERE array_length(destinations, 1) > 0;

-- Drop destinations column
ALTER TABLE reservations DROP COLUMN IF EXISTS destinations;
