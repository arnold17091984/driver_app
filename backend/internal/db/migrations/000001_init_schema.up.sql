-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- 1. users
-- ============================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     VARCHAR(50)  NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    role            VARCHAR(20)  NOT NULL CHECK (role IN ('admin','dispatcher','viewer','driver')),
    priority_level  INTEGER      NOT NULL DEFAULT 0,
    fcm_token       TEXT,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_employee_id ON users(employee_id);
CREATE INDEX idx_users_role ON users(role);

-- ============================================
-- 2. vehicles
-- ============================================
CREATE TABLE vehicles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    license_plate   VARCHAR(20)  NOT NULL UNIQUE,
    driver_id       UUID         NOT NULL REFERENCES users(id),
    is_maintenance  BOOLEAN      NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vehicles_driver_id ON vehicles(driver_id);

-- ============================================
-- 3. driver_attendance
-- ============================================
CREATE TABLE driver_attendance (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id       UUID         NOT NULL REFERENCES users(id),
    clock_in_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    clock_out_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attendance_driver_id ON driver_attendance(driver_id);
CREATE INDEX idx_attendance_clock_in ON driver_attendance(clock_in_at);
CREATE INDEX idx_attendance_active ON driver_attendance(driver_id) WHERE clock_out_at IS NULL;

-- ============================================
-- 4. vehicle_locations (history)
-- ============================================
CREATE TABLE vehicle_locations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id      UUID         NOT NULL REFERENCES vehicles(id),
    location        GEOGRAPHY(POINT, 4326) NOT NULL,
    heading         REAL,
    speed           REAL,
    accuracy        REAL,
    recorded_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vehicle_locations_vehicle_id ON vehicle_locations(vehicle_id);
CREATE INDEX idx_vehicle_locations_recorded_at ON vehicle_locations(recorded_at DESC);
CREATE INDEX idx_vehicle_locations_geo ON vehicle_locations USING GIST(location);

-- ============================================
-- 5. vehicle_location_current (latest known)
-- ============================================
CREATE TABLE vehicle_location_current (
    vehicle_id      UUID PRIMARY KEY REFERENCES vehicles(id),
    location        GEOGRAPHY(POINT, 4326) NOT NULL,
    heading         REAL,
    speed           REAL,
    accuracy        REAL,
    recorded_at     TIMESTAMPTZ  NOT NULL,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vlc_geo ON vehicle_location_current USING GIST(location);

-- ============================================
-- 6. dispatches
-- ============================================
CREATE TABLE dispatches (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id             UUID         REFERENCES vehicles(id),
    requester_id           UUID         NOT NULL REFERENCES users(id),
    dispatcher_id          UUID         REFERENCES users(id),
    purpose                TEXT         NOT NULL,
    passenger_name         VARCHAR(200),
    passenger_count        INTEGER      DEFAULT 1,
    notes                  TEXT,
    pickup_address         TEXT         NOT NULL,
    pickup_location        GEOGRAPHY(POINT, 4326),
    dropoff_address        TEXT,
    dropoff_location       GEOGRAPHY(POINT, 4326),
    status                 VARCHAR(20)  NOT NULL DEFAULT 'pending'
                           CHECK (status IN ('pending','assigned','accepted','en_route',
                                             'arrived','completed','cancelled')),
    estimated_duration_sec INTEGER,
    estimated_distance_m   INTEGER,
    assigned_at            TIMESTAMPTZ,
    accepted_at            TIMESTAMPTZ,
    en_route_at            TIMESTAMPTZ,
    arrived_at             TIMESTAMPTZ,
    completed_at           TIMESTAMPTZ,
    cancelled_at           TIMESTAMPTZ,
    cancel_reason          TEXT,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dispatches_vehicle_id ON dispatches(vehicle_id);
CREATE INDEX idx_dispatches_status ON dispatches(status);
CREATE INDEX idx_dispatches_created_at ON dispatches(created_at DESC);
CREATE INDEX idx_dispatches_active ON dispatches(vehicle_id, status)
    WHERE status IN ('assigned','accepted','en_route','arrived');

-- ============================================
-- 7. dispatch_eta_snapshots
-- ============================================
CREATE TABLE dispatch_eta_snapshots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispatch_id     UUID         NOT NULL REFERENCES dispatches(id) ON DELETE CASCADE,
    vehicle_id      UUID         NOT NULL REFERENCES vehicles(id),
    duration_sec    INTEGER      NOT NULL,
    distance_m      INTEGER      NOT NULL,
    origin_lat      DOUBLE PRECISION NOT NULL,
    origin_lng      DOUBLE PRECISION NOT NULL,
    calculated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_eta_dispatch ON dispatch_eta_snapshots(dispatch_id);

-- ============================================
-- 8. reservations
-- ============================================
CREATE TABLE reservations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id      UUID         NOT NULL REFERENCES vehicles(id),
    requester_id    UUID         NOT NULL REFERENCES users(id),
    start_time      TIMESTAMPTZ  NOT NULL,
    end_time        TIMESTAMPTZ  NOT NULL,
    purpose         TEXT         NOT NULL,
    destination     TEXT,
    notes           TEXT,
    priority_level  INTEGER      NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'confirmed'
                    CHECK (status IN ('confirmed','pending_conflict','cancelled','completed')),
    cancel_reason   TEXT,
    cancelled_by    UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_time_range CHECK (end_time > start_time)
);

CREATE INDEX idx_reservations_vehicle_id ON reservations(vehicle_id);
CREATE INDEX idx_reservations_time ON reservations(vehicle_id, start_time, end_time);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_reservations_conflict ON reservations(status) WHERE status = 'pending_conflict';
CREATE INDEX idx_reservations_overlap ON reservations(vehicle_id, start_time, end_time)
    WHERE status IN ('confirmed', 'pending_conflict');

-- ============================================
-- 9. reservation_conflicts
-- ============================================
CREATE TABLE reservation_conflicts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    winning_reservation_id  UUID NOT NULL REFERENCES reservations(id),
    losing_reservation_id   UUID NOT NULL REFERENCES reservations(id),
    status                  VARCHAR(20) NOT NULL DEFAULT 'pending'
                            CHECK (status IN ('pending','resolved_reassign','resolved_changed',
                                              'resolved_cancelled','force_assigned')),
    resolved_by             UUID REFERENCES users(id),
    resolution_reason       TEXT,
    resolved_at             TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conflicts_status ON reservation_conflicts(status);
CREATE INDEX idx_conflicts_losing ON reservation_conflicts(losing_reservation_id);

-- ============================================
-- 10. audit_logs
-- ============================================
CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id        UUID         NOT NULL REFERENCES users(id),
    action          VARCHAR(50)  NOT NULL,
    target_type     VARCHAR(50)  NOT NULL,
    target_id       UUID         NOT NULL,
    before_state    JSONB,
    after_state     JSONB,
    reason          TEXT,
    ip_address      INET,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_target ON audit_logs(target_type, target_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);

-- ============================================
-- 11. notification_logs
-- ============================================
CREATE TABLE notification_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users(id),
    type            VARCHAR(50)  NOT NULL,
    title           VARCHAR(200) NOT NULL,
    body            TEXT         NOT NULL,
    data            JSONB,
    sent_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    delivered       BOOLEAN      DEFAULT false,
    fcm_message_id  VARCHAR(200)
);

CREATE INDEX idx_notif_user ON notification_logs(user_id);
CREATE INDEX idx_notif_type ON notification_logs(type);
