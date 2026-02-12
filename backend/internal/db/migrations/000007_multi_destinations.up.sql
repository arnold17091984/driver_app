-- Add destinations TEXT[] column to support multiple destinations (e.g. delivery routes)
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS destinations TEXT[] DEFAULT '{}';

-- Migrate existing destination data to destinations array
UPDATE reservations SET destinations = ARRAY[destination] WHERE destination IS NOT NULL AND destination != '';

-- Drop old destination column
ALTER TABLE reservations DROP COLUMN IF EXISTS destination;
