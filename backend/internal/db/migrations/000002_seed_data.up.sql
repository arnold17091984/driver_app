-- Seed data for development ONLY.
-- WARNING: These users use a well-known password and must NOT exist in production.
-- The application will refuse to start in production if these users are active.
-- Password for all users: "password123" (bcrypt hash)
-- $2a$10$XYof2X4G00pN9LmpwyM/z.nuHsRZFIN23/JRJ9gNHdSRSHipwQRVe

-- Admin user
INSERT INTO users (employee_id, password_hash, name, role, priority_level) VALUES
    ('admin001', '$2a$10$XYof2X4G00pN9LmpwyM/z.nuHsRZFIN23/JRJ9gNHdSRSHipwQRVe', 'Admin Garcia', 'admin', 10);

-- Dispatcher
INSERT INTO users (employee_id, password_hash, name, role, priority_level) VALUES
    ('dispatch001', '$2a$10$XYof2X4G00pN9LmpwyM/z.nuHsRZFIN23/JRJ9gNHdSRSHipwQRVe', 'Maria Santos', 'dispatcher', 7);

-- Viewer
INSERT INTO users (employee_id, password_hash, name, role, priority_level) VALUES
    ('viewer001', '$2a$10$XYof2X4G00pN9LmpwyM/z.nuHsRZFIN23/JRJ9gNHdSRSHipwQRVe', 'Jose Reyes', 'viewer', 3);

-- 5 Drivers
INSERT INTO users (employee_id, password_hash, name, role, priority_level) VALUES
    ('driver001', '$2a$10$XYof2X4G00pN9LmpwyM/z.nuHsRZFIN23/JRJ9gNHdSRSHipwQRVe', 'Carlo Dela Cruz', 'driver', 1),
    ('driver002', '$2a$10$XYof2X4G00pN9LmpwyM/z.nuHsRZFIN23/JRJ9gNHdSRSHipwQRVe', 'Marco Villanueva', 'driver', 1),
    ('driver003', '$2a$10$XYof2X4G00pN9LmpwyM/z.nuHsRZFIN23/JRJ9gNHdSRSHipwQRVe', 'Rodel Bautista', 'driver', 1),
    ('driver004', '$2a$10$XYof2X4G00pN9LmpwyM/z.nuHsRZFIN23/JRJ9gNHdSRSHipwQRVe', 'Jay Mendoza', 'driver', 1),
    ('driver005', '$2a$10$XYof2X4G00pN9LmpwyM/z.nuHsRZFIN23/JRJ9gNHdSRSHipwQRVe', 'Ariel Gonzales', 'driver', 1);

-- 5 Vehicles (assigned to drivers)
INSERT INTO vehicles (name, license_plate, driver_id) VALUES
    ('Van A', 'NCR-1001', (SELECT id FROM users WHERE employee_id = 'driver001')),
    ('Van B', 'NCR-1002', (SELECT id FROM users WHERE employee_id = 'driver002')),
    ('Sedan C', 'NCR-1003', (SELECT id FROM users WHERE employee_id = 'driver003')),
    ('SUV D', 'NCR-1004', (SELECT id FROM users WHERE employee_id = 'driver004')),
    ('Van E', 'NCR-1005', (SELECT id FROM users WHERE employee_id = 'driver005'));
