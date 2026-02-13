CREATE INDEX IF NOT EXISTS idx_reservations_requester_id ON reservations(requester_id);
CREATE INDEX IF NOT EXISTS idx_dispatches_requester_id ON dispatches(requester_id);
CREATE INDEX IF NOT EXISTS idx_dispatches_dispatcher_id ON dispatches(dispatcher_id);
CREATE INDEX IF NOT EXISTS idx_users_role_active ON users(role, is_active) WHERE is_active = true;
