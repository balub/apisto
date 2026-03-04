-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Projects table
CREATE TABLE projects (
    id TEXT PRIMARY KEY DEFAULT encode(gen_random_bytes(8), 'hex'),
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Devices table
CREATE TABLE devices (
    id TEXT PRIMARY KEY DEFAULT encode(gen_random_bytes(8), 'hex'),
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL DEFAULT encode(gen_random_bytes(16), 'hex'),
    name TEXT NOT NULL DEFAULT 'Unnamed Device',
    description TEXT DEFAULT '',
    metadata JSONB DEFAULT '{}',
    is_online BOOLEAN DEFAULT FALSE,
    last_seen_at TIMESTAMPTZ,
    first_seen_at TIMESTAMPTZ,
    firmware_version TEXT DEFAULT '',
    ip_address TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_devices_project ON devices(project_id);
CREATE INDEX idx_devices_token ON devices(token);

-- Telemetry table (TimescaleDB hypertable)
CREATE TABLE telemetry (
    time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    device_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value_numeric DOUBLE PRECISION,
    value_text TEXT,
    value_bool BOOLEAN,
    value_type TEXT NOT NULL CHECK (value_type IN ('number', 'string', 'boolean', 'json'))
);

SELECT create_hypertable('telemetry', 'time', chunk_time_interval => INTERVAL '1 day');

CREATE INDEX idx_telemetry_device_key_time ON telemetry(device_id, key, time DESC);
CREATE INDEX idx_telemetry_device_time ON telemetry(device_id, time DESC);

-- Device keys (tracks discovered data keys per device for auto-dashboard)
CREATE TABLE device_keys (
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value_type TEXT NOT NULL,
    first_seen_at TIMESTAMPTZ DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ DEFAULT NOW(),
    widget_type TEXT DEFAULT 'auto',
    display_name TEXT DEFAULT '',
    unit TEXT DEFAULT '',
    sort_order INTEGER DEFAULT 0,
    PRIMARY KEY (device_id, key)
);

-- Commands table (cloud-to-device)
CREATE TABLE commands (
    id TEXT PRIMARY KEY DEFAULT encode(gen_random_bytes(8), 'hex'),
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    command TEXT NOT NULL,
    payload TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'acknowledged', 'failed', 'expired')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    sent_at TIMESTAMPTZ,
    acked_at TIMESTAMPTZ
);

CREATE INDEX idx_commands_device ON commands(device_id, created_at DESC);

-- Dashboard shares
CREATE TABLE dashboard_shares (
    id TEXT PRIMARY KEY DEFAULT encode(gen_random_bytes(8), 'hex'),
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    share_token TEXT UNIQUE NOT NULL DEFAULT encode(gen_random_bytes(16), 'hex'),
    is_active BOOLEAN DEFAULT TRUE,
    password_hash TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
