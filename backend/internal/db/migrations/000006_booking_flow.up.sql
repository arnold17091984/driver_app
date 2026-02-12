-- Add pending_driver and driver_declined statuses to reservations
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS reservations_status_check;
ALTER TABLE reservations ADD CONSTRAINT reservations_status_check
  CHECK (status IN ('confirmed', 'pending_conflict', 'pending_driver', 'driver_declined', 'cancelled', 'completed'));

-- New columns for unified booking flow
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS passenger_name VARCHAR(200);
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS pickup_address TEXT;
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS pickup_location GEOGRAPHY(POINT, 4326);
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS declined_by_driver_ids UUID[] DEFAULT '{}';

-- Index for driver polling of pending reservations
CREATE INDEX IF NOT EXISTS idx_reservations_pending_driver
  ON reservations(vehicle_id, status) WHERE status = 'pending_driver';
