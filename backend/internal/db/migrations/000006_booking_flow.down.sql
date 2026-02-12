DROP INDEX IF EXISTS idx_reservations_pending_driver;

ALTER TABLE reservations DROP COLUMN IF EXISTS declined_by_driver_ids;
ALTER TABLE reservations DROP COLUMN IF EXISTS pickup_location;
ALTER TABLE reservations DROP COLUMN IF EXISTS pickup_address;
ALTER TABLE reservations DROP COLUMN IF EXISTS passenger_name;

ALTER TABLE reservations DROP CONSTRAINT IF EXISTS reservations_status_check;
ALTER TABLE reservations ADD CONSTRAINT reservations_status_check
  CHECK (status IN ('confirmed', 'pending_conflict', 'cancelled', 'completed'));
