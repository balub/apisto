# Apisto v1 — Claude Code Build Prompt

## What is Apisto?

Apisto is a self-hosted, lightweight, hardware-agnostic IoT backend platform. Think "Firebase for IoT" — a single self-hosted server that lets hardware makers connect devices, stream telemetry, visualize data in real-time, and control devices remotely. The primary audience is hobbyists and makers using ESP32/ESP8266 boards.

The single most important design goal: **a maker with an ESP32 should go from `docker compose up` to a live interactive dashboard in under 15 minutes.**

---

## Tech Stack

- **Language:** Go (for the server — single binary, easy cross-compilation, great concurrency, low memory footprint)
- **MQTT Broker:** Eclipse Mosquitto (external, managed via Docker Compose). Battle-tested, lightweight, MQTT v5 compliant. The Apisto Go server connects to Mosquitto as an MQTT client to subscribe to device topics and process incoming telemetry. Mosquitto handles all broker responsibilities (connection management, QoS, retained messages, session persistence). Configure Mosquitto with a custom auth plugin or use its `mosquitto_passwd` file for device token auth — OR have the Apisto server act as a bridge where Mosquitto allows all connections and the Apisto MQTT subscriber validates tokens at the application layer and ignores messages from unregistered devices.
- **Database (Relational):** PostgreSQL 16 for device metadata, projects, commands, dashboard shares, and all relational data. Managed via Docker Compose.
- **Database (Time-Series):** TimescaleDB (PostgreSQL extension) for all telemetry data. This runs inside the SAME PostgreSQL instance — TimescaleDB extends Postgres with hypertables, automatic time-based partitioning (chunking), built-in compression, continuous aggregates, and retention policies. This means we use ONE database for everything (relational + time-series), queried with standard SQL. Do NOT use InfluxDB or a separate TSDB — TimescaleDB gives us the best of both worlds in a single Postgres instance. Use the `timescale/timescaledb:latest-pg16` Docker image which bundles Postgres + TimescaleDB.
- **Database Migrations:** Use [golang-migrate](https://github.com/golang-migrate/migrate) for versioned SQL migrations.
- **Web Framework:** [Fiber](https://github.com/gofiber/fiber) or [Echo](https://github.com/labstack/echo) for the REST API and WebSocket support.
- **Frontend:** React with Vite. Tailwind CSS for styling. Recharts for auto-generated charts. The frontend is a SPA. In development, Vite runs on its own port with proxy to the Go API. In production, the built frontend is either served by the Go server using `embed` OR by a lightweight nginx container — either approach works.
- **WebSocket:** For real-time dashboard updates. When a device sends telemetry via MQTT, the Go server's MQTT subscriber receives it, stores it, and pushes it to connected dashboard clients over WebSocket.
- **Containerization:** Docker Compose with separate services for: Apisto API server (Go), TimescaleDB/Postgres, and Mosquitto. The full stack spins up with a single `docker compose up`. Must run on a Raspberry Pi 4 (ARM64) and any $5 VPS with 1GB RAM.

---

## Project Structure

```
apisto/
├── cmd/
│   └── apisto/
│       └── main.go              # Entry point
├── internal/
│   ├── server/
│   │   └── server.go            # Main server orchestrator (starts HTTP, MQTT subscriber, WebSocket)
│   ├── config/
│   │   └── config.go            # Configuration loading (env vars, config file, defaults)
│   ├── database/
│   │   ├── postgres.go          # PostgreSQL/TimescaleDB connection, migration runner
│   │   └── migrations/          # Versioned SQL migration files (001_init.up.sql, etc.)
│   ├── mqtt/
│   │   └── client.go            # MQTT client that connects to Mosquitto, subscribes to topics, processes messages
│   ├── api/
│   │   ├── router.go            # Route definitions
│   │   ├── middleware.go         # Auth middleware, CORS, logging
│   │   ├── projects.go          # Project CRUD handlers
│   │   ├── devices.go           # Device CRUD + status handlers
│   │   ├── telemetry.go         # Telemetry ingestion (HTTP) + query handlers
│   │   ├── commands.go          # Cloud-to-device command handlers
│   │   └── websocket.go         # WebSocket upgrade + real-time push
│   ├── models/
│   │   ├── project.go
│   │   ├── device.go
│   │   ├── telemetry.go
│   │   └── command.go
│   ├── services/
│   │   ├── device_service.go    # Device registration, status tracking, heartbeat
│   │   ├── telemetry_service.go # Telemetry ingestion, storage, querying, aggregations (TimescaleDB)
│   │   ├── command_service.go   # Command dispatch via MQTT publish, acknowledgment tracking
│   │   └── realtime_service.go  # WebSocket hub, fan-out to dashboard clients
│   └── auth/
│       └── token.go             # Token generation, validation
├── web/                         # React frontend (Vite)
│   ├── src/
│   │   ├── App.jsx
│   │   ├── main.jsx
│   │   ├── pages/
│   │   │   ├── Dashboard.jsx     # Main project dashboard
│   │   │   ├── DeviceView.jsx    # Single device dashboard with auto-widgets
│   │   │   ├── Projects.jsx      # Project list
│   │   │   ├── QuickStart.jsx    # Getting started page with code snippets
│   │   │   └── Settings.jsx      # Server/project settings
│   │   ├── components/
│   │   │   ├── AutoDashboard.jsx # Auto-generated widget layout
│   │   │   ├── widgets/
│   │   │   │   ├── LineChart.jsx
│   │   │   │   ├── GaugeCard.jsx
│   │   │   │   ├── BooleanIndicator.jsx
│   │   │   │   ├── ValueCard.jsx
│   │   │   │   ├── EventLog.jsx
│   │   │   │   └── ControlWidget.jsx  # Toggle, slider, text input for commands
│   │   │   ├── DeviceStatusBadge.jsx
│   │   │   ├── Sidebar.jsx
│   │   │   └── CodeSnippet.jsx   # Syntax-highlighted, copy-to-clipboard code block
│   │   ├── hooks/
│   │   │   ├── useWebSocket.js   # WebSocket connection + auto-reconnect
│   │   │   └── useDeviceData.js  # Fetch + stream device telemetry
│   │   └── lib/
│   │       └── api.js            # API client
│   ├── index.html
│   ├── vite.config.js
│   ├── tailwind.config.js
│   └── package.json
├── sdk/
│   └── arduino/
│       ├── Apisto.h              # Single-header Arduino library
│       └── Apisto.cpp
│       └── examples/
│           ├── basic_telemetry/
│           │   └── basic_telemetry.ino
│           ├── control/
│           │   └── control.ino
│           └── full_example/
│               └── full_example.ino
├── mosquitto/
│   └── mosquitto.conf            # Mosquitto broker configuration
├── Dockerfile                    # Apisto Go server only
├── docker-compose.yml            # Full stack: apisto + timescaledb + mosquitto
├── Makefile
├── go.mod
├── go.sum
├── README.md
└── .env.example
```

---

## Database Schema (PostgreSQL + TimescaleDB)

All tables live in a single PostgreSQL instance with TimescaleDB extension enabled.

### Migration 001: Init

```sql
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
-- This is the core time-series table. TimescaleDB automatically partitions it
-- into time-based chunks for fast inserts and time-range queries.
CREATE TABLE telemetry (
    time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    device_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value_numeric DOUBLE PRECISION,
    value_text TEXT,
    value_bool BOOLEAN,
    value_type TEXT NOT NULL CHECK (value_type IN ('number', 'string', 'boolean', 'json'))
);

-- Convert to TimescaleDB hypertable — partitioned by time with 1-day chunks
-- This is the magic: inserts are fast, time-range queries are fast, old data
-- can be compressed or dropped by chunk.
SELECT create_hypertable('telemetry', 'time', chunk_time_interval => INTERVAL '1 day');

-- Indexes optimized for common query patterns
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
```

### Migration 002: TimescaleDB Retention Policy + Continuous Aggregates

```sql
-- Automatic data retention: drop telemetry chunks older than 30 days
-- This is a TimescaleDB background job — no cron needed, no manual cleanup.
SELECT add_retention_policy('telemetry', INTERVAL '30 days');

-- Enable compression on chunks older than 7 days
-- This dramatically reduces storage for historical data.
ALTER TABLE telemetry SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id, key',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('telemetry', INTERVAL '7 days');

-- Continuous aggregate: pre-compute hourly averages for fast dashboard queries
-- TimescaleDB materializes this in the background — querying hourly data
-- becomes a simple table read instead of aggregating raw rows.
CREATE MATERIALIZED VIEW telemetry_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    device_id,
    key,
    AVG(value_numeric) AS avg_value,
    MIN(value_numeric) AS min_value,
    MAX(value_numeric) AS max_value,
    COUNT(*) AS sample_count,
    last(value_numeric, time) AS last_value
FROM telemetry
WHERE value_type = 'number'
GROUP BY bucket, device_id, key
WITH NO DATA;

-- Refresh the continuous aggregate automatically
SELECT add_continuous_aggregate_policy('telemetry_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour'
);

-- Optional: daily aggregate for long-range dashboards
CREATE MATERIALIZED VIEW telemetry_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    device_id,
    key,
    AVG(value_numeric) AS avg_value,
    MIN(value_numeric) AS min_value,
    MAX(value_numeric) AS max_value,
    COUNT(*) AS sample_count,
    last(value_numeric, time) AS last_value
FROM telemetry
WHERE value_type = 'number'
GROUP BY bucket, device_id, key
WITH NO DATA;

SELECT add_continuous_aggregate_policy('telemetry_daily',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day'
);
```

---

## MQTT Topic Structure

Devices communicate over MQTT using this topic hierarchy:

```
# Device publishes telemetry
apisto/{device_token}/telemetry

# Device publishes heartbeat/status
apisto/{device_token}/status

# Server publishes commands TO device (device subscribes)
apisto/{device_token}/commands

# Device publishes command acknowledgments
apisto/{device_token}/commands/ack
```

### MQTT Auth

Mosquitto is configured with `allow_anonymous true` for v1 simplicity. Authentication is handled at the application layer:
- The Apisto Go server subscribes to `apisto/+/telemetry`, `apisto/+/status`, and `apisto/+/commands/ack` using a wildcard subscription
- When a message arrives, the server extracts the device token from the topic
- The server validates the token against the devices table in Postgres
- If the token is invalid, the message is silently dropped (logged at debug level)
- On first valid message from an unregistered device: auto-register it if the token matches a valid device in the DB
- For v2, consider upgrading to Mosquitto's `mosquitto-go-auth` plugin for broker-level auth with a Postgres backend

### Telemetry Payload Format

Devices send JSON payloads:

```json
{
  "temperature": 24.5,
  "humidity": 60,
  "relay_on": true,
  "status_msg": "all good"
}
```

The server automatically:
1. Parses each key-value pair
2. Detects the type (number, boolean, string)
3. Stores in the telemetry table with appropriate value column
4. Updates the device_keys table (upserts key metadata)
5. Pushes to WebSocket subscribers for that device

---

## REST API Endpoints

### Projects
```
POST   /api/v1/projects                    # Create project
GET    /api/v1/projects                    # List projects
GET    /api/v1/projects/:id                # Get project
PUT    /api/v1/projects/:id                # Update project
DELETE /api/v1/projects/:id                # Delete project
```

### Devices
```
POST   /api/v1/projects/:id/devices        # Create device (generates token)
GET    /api/v1/projects/:id/devices        # List devices in project
GET    /api/v1/devices/:id                 # Get device details + status
PUT    /api/v1/devices/:id                 # Update device metadata
DELETE /api/v1/devices/:id                 # Delete device
GET    /api/v1/devices/:id/keys            # Get discovered data keys for device
```

### Telemetry
```
POST   /api/v1/devices/:token/telemetry    # HTTP telemetry ingestion (alternative to MQTT)
GET    /api/v1/devices/:id/telemetry       # Query telemetry data
       ?key=temperature                     # Filter by key
       &from=2024-01-01T00:00:00Z          # Start time
       &to=2024-01-02T00:00:00Z            # End time
       &limit=100                           # Max rows
       &order=desc                          # Sort order
       &aggregate=avg                       # Aggregation: avg, min, max, count, last
       &interval=1h                         # Aggregation window: 1m, 5m, 15m, 1h, 1d
GET    /api/v1/devices/:id/telemetry/latest # Get latest value for each key
```

### Commands
```
POST   /api/v1/devices/:id/commands        # Send command to device
       Body: { "command": "relay", "payload": "on" }
GET    /api/v1/devices/:id/commands        # List command history
```

### Dashboard Shares
```
POST   /api/v1/devices/:id/share           # Create share link
DELETE /api/v1/shares/:share_token         # Revoke share
GET    /api/v1/public/:share_token         # Public device dashboard data (no auth)
GET    /api/v1/public/:share_token/ws      # Public WebSocket stream (no auth)
```

### WebSocket
```
GET    /api/v1/devices/:id/ws              # WebSocket stream for device telemetry + status
```

WebSocket message format (server → client):
```json
{
  "type": "telemetry",
  "device_id": "abc123",
  "data": { "temperature": 24.5, "humidity": 60 },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

```json
{
  "type": "status",
  "device_id": "abc123",
  "is_online": true,
  "last_seen_at": "2024-01-15T10:30:00Z"
}
```

```json
{
  "type": "command_ack",
  "command_id": "cmd_xyz",
  "status": "acknowledged"
}
```

---

## Server Architecture Details

### Main Server Orchestrator (`internal/server/server.go`)

The server starts all subsystems in a single process:

1. **Load config** from env vars / config file / defaults
2. **Connect to PostgreSQL/TimescaleDB** — run migrations via golang-migrate, set up connection pool
3. **Connect to Mosquitto as MQTT client** using [paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) — subscribe to wildcard topics (`apisto/+/telemetry`, `apisto/+/status`, `apisto/+/commands/ack`), register message handlers
4. **Start HTTP/WebSocket server** on configurable port (default: 8080) with embedded or proxied React frontend
5. **Start background workers:**
   - Heartbeat checker: every 30s, mark devices as offline if no activity for configurable timeout (default: 60s)
   - TimescaleDB handles data retention and compression automatically via policies — no manual pruning needed

### Config (`internal/config/config.go`)

All configurable via environment variables with sensible defaults:

```
APISTO_PORT=8080                            # HTTP/WebSocket port
APISTO_DATABASE_URL=postgres://apisto:apisto@timescaledb:5432/apisto?sslmode=disable
APISTO_MQTT_BROKER_URL=tcp://mosquitto:1883 # Mosquitto connection URL
APISTO_MQTT_CLIENT_ID=apisto-server         # MQTT client ID for the server
APISTO_RETENTION_DAYS=30                    # Telemetry retention (applied via TimescaleDB policy)
APISTO_HEARTBEAT_TIMEOUT=60                 # Seconds before device marked offline
APISTO_LOG_LEVEL=info                       # Log level
APISTO_CORS_ORIGINS=*                       # CORS origins
```

### MQTT Client (`internal/mqtt/client.go`)

The Go server connects to Mosquitto as an MQTT client (NOT a broker):

- Use `paho.mqtt.golang` library
- Connect to Mosquitto with `APISTO_MQTT_BROKER_URL`
- Subscribe to `apisto/+/telemetry` with QoS 1
- Subscribe to `apisto/+/status` with QoS 1
- Subscribe to `apisto/+/commands/ack` with QoS 1
- On telemetry message:
  - Extract device token from topic (second segment)
  - Validate token against DB (cache valid tokens in memory with TTL for performance)
  - Parse JSON payload
  - Call telemetry service to store data in TimescaleDB
  - Fan out to WebSocket hub for real-time dashboard updates
- On status message:
  - Update device online status and last_seen_at
- On command ack:
  - Update command status to "acknowledged"
- Handle disconnects with automatic reconnection (built into paho client)
- When sending commands TO a device (triggered by REST API):
  - Publish to `apisto/{device_token}/commands` topic via the same MQTT client

### Real-time Service (`internal/services/realtime_service.go`)

WebSocket hub pattern:
- Map of `device_id → set of WebSocket connections`
- When telemetry arrives (from MQTT hook or HTTP endpoint), fan out to all connected WebSocket clients for that device
- Handle connection/disconnection cleanly
- Support public share WebSocket connections (validated via share token)

### Telemetry Service (`internal/services/telemetry_service.go`)

- **Ingest:** Parse JSON payload, detect types, batch insert into telemetry hypertable, upsert device_keys
- **Query:** Support time-range queries with optional key filter, ordering, limit
- **Aggregate:** Use TimescaleDB's `time_bucket` function for blazing-fast aggregations:
  ```sql
  SELECT
    time_bucket('1 hour', time) AS bucket,
    AVG(value_numeric) AS value
  FROM telemetry
  WHERE device_id = $1 AND key = $2 AND time BETWEEN $3 AND $4
  GROUP BY bucket
  ORDER BY bucket ASC
  ```
  For longer time ranges (>24h), query the `telemetry_hourly` continuous aggregate instead of raw data:
  ```sql
  SELECT bucket, avg_value AS value, min_value, max_value, sample_count
  FROM telemetry_hourly
  WHERE device_id = $1 AND key = $2 AND bucket BETWEEN $3 AND $4
  ORDER BY bucket ASC
  ```
  For ranges >7 days, use `telemetry_daily`.
- **Latest:** Return the most recent value for each key for a device using TimescaleDB's `last()` aggregate or `DISTINCT ON`:
  ```sql
  SELECT DISTINCT ON (key)
    key, value_numeric, value_text, value_bool, value_type, time
  FROM telemetry
  WHERE device_id = $1
  ORDER BY key, time DESC
  ```

---

## Frontend Details

### Tech
- React 18+ with Vite
- Tailwind CSS (keep it clean, minimal, dark mode default)
- Recharts for charts
- React Router for navigation
- Native WebSocket (no Socket.io — keep it light)

### Design Language
- Dark mode default, optional light mode
- Clean, minimal, developer-focused
- Monospace fonts for data/values
- Color palette: dark grays (#0a0a0a, #1a1a1a, #2a2a2a) with a vibrant accent color (teal/cyan #06b6d4 or green #10b981)
- Card-based layout for widgets
- Status indicators: green dot = online, gray dot = offline
- NO enterprise look. Think Linear/Vercel/Raycast aesthetic

### Pages

**Projects Page (`/`)**
- Card grid showing all projects
- Each card shows: project name, device count, last activity
- "New Project" button → modal with name + description
- Click project → goes to project dashboard

**Project Dashboard (`/projects/:id`)**
- Left sidebar: list of devices with online/offline indicators
- Main area: selected device's auto-dashboard
- Top bar: project name, settings gear, "Add Device" button
- "Add Device" → generates token, shows Quick Start code snippet

**Device View (`/projects/:id/devices/:deviceId`)**
- Top: device name, status badge (online/offline), last seen, IP, firmware version, edit button
- Middle: Auto-Dashboard grid of widgets
- Bottom: Control panel (command widgets) + Command history log
- Tab: Raw telemetry data table with filters

**Quick Start Page (`/quickstart`)**
- Step-by-step visual guide
- Step 1: Start server (docker run command, copy-to-clipboard)
- Step 2: Create project (done via UI)
- Step 3: Add device (shows generated token)
- Step 4: Upload sketch (Arduino code pre-filled with token and server IP, copy-to-clipboard with syntax highlighting)
- Step 5: See data (animated arrow pointing to dashboard)
- The Arduino code snippet must be pre-filled with the actual device token and server URL

**Settings Page (`/settings`)**
- Data retention configuration
- Heartbeat timeout
- MQTT port display
- Server info (version, uptime, storage usage)

### Auto-Dashboard Widget Logic

The auto-dashboard is the hero feature. It auto-generates widgets based on telemetry data:

```
Widget Selection Rules:
1. Numeric value → ValueCard (current value + sparkline of last 20 points)
2. Boolean value → BooleanIndicator (on/off with color: green/red)
3. String value → EventLog (scrolling list of recent values with timestamps)
4. Key named "lat"/"lng" or "latitude"/"longitude" → Map widget (stretch goal, skip in v1 if complex)
5. Any numeric key with >10 data points → also gets a LineChart widget
```

Layout:
- CSS Grid, responsive
- 2 columns on desktop, 1 on mobile
- ValueCards are 1 column wide
- LineCharts span full width
- BooleanIndicators are 1 column wide
- EventLogs span full width

Each widget shows:
- Key name as title (with editable display name)
- Unit label (if configured)
- Last updated timestamp
- Real-time updates via WebSocket (value animates/transitions when new data arrives)

### Control Widgets

Displayed in a separate "Controls" section below the auto-dashboard:

- Created manually by user via "Add Control" button
- Types: Toggle (sends "on"/"off"), Slider (sends numeric value), Button (sends configurable payload), Text Input (sends string)
- Each control maps to a command name
- When activated, sends POST to `/api/v1/devices/:id/commands`
- Shows last sent status (pending → sent → acknowledged)

---

## Arduino SDK (`sdk/arduino/`)

### Design Principles
- Single `.h` and `.cpp` file — no complex library structure
- Depends only on: WiFi.h (ESP32) or ESP8266WiFi.h, PubSubClient (MQTT), ArduinoJson
- Connects to Apisto server via MQTT
- Simple, chainable API
- Auto-reconnect on WiFi/MQTT disconnect
- Non-blocking (no delay() calls internally)

### API Surface

```cpp
#include <Apisto.h>

// Constructor
Apisto device(DEVICE_TOKEN, SERVER_HOST, MQTT_PORT);

// Initialize connection (call in setup())
void device.begin(const char* wifiSSID, const char* wifiPassword);

// Send telemetry (call in loop())
void device.send(const char* key, float value);
void device.send(const char* key, int value);
void device.send(const char* key, bool value);
void device.send(const char* key, const char* value);

// Send multiple values at once (batched in single MQTT message)
void device.beginBatch();
void device.add(const char* key, float value);
void device.add(const char* key, bool value);
// ... etc
void device.endBatch();  // Publishes all added values as single JSON

// Register command handler
void device.onCommand(const char* command, void (*callback)(String payload));

// Must be called in loop()
void device.loop();

// Status
bool device.isConnected();
```

### Example Sketches

**basic_telemetry.ino:**
```cpp
#include <Apisto.h>
#include <DHT.h>

#define DHT_PIN 4
#define DEVICE_TOKEN "your_device_token"
#define SERVER_HOST "192.168.1.100"

DHT dht(DHT_PIN, DHT22);
Apisto device(DEVICE_TOKEN, SERVER_HOST);

void setup() {
  Serial.begin(115200);
  dht.begin();
  device.begin("YourWiFi", "YourPassword");
}

void loop() {
  device.loop();

  static unsigned long lastSend = 0;
  if (millis() - lastSend > 5000) {
    device.send("temperature", dht.readTemperature());
    device.send("humidity", dht.readHumidity());
    lastSend = millis();
  }
}
```

**control.ino:**
```cpp
#include <Apisto.h>

#define RELAY_PIN 5
#define LED_PIN 2
#define DEVICE_TOKEN "your_device_token"
#define SERVER_HOST "192.168.1.100"

Apisto device(DEVICE_TOKEN, SERVER_HOST);

void setup() {
  Serial.begin(115200);
  pinMode(RELAY_PIN, OUTPUT);
  pinMode(LED_PIN, OUTPUT);

  device.begin("YourWiFi", "YourPassword");

  device.onCommand("relay", [](String payload) {
    digitalWrite(RELAY_PIN, payload == "on" ? HIGH : LOW);
    Serial.println("Relay: " + payload);
  });

  device.onCommand("led", [](String payload) {
    int brightness = payload.toInt();
    analogWrite(LED_PIN, brightness);
    Serial.println("LED brightness: " + payload);
  });
}

void loop() {
  device.loop();

  static unsigned long lastSend = 0;
  if (millis() - lastSend > 5000) {
    device.send("relay_state", digitalRead(RELAY_PIN) == HIGH);
    device.send("uptime_seconds", (int)(millis() / 1000));
    lastSend = millis();
  }
}
```

---

## Docker Compose Setup

The full stack runs with `docker compose up`. Three services: the Apisto Go server, TimescaleDB (Postgres), and Mosquitto.

### docker-compose.yml

```yaml
version: "3.8"

services:
  # TimescaleDB (PostgreSQL + TimescaleDB extension)
  timescaledb:
    image: timescale/timescaledb:latest-pg16
    container_name: apisto-db
    restart: unless-stopped
    environment:
      POSTGRES_USER: apisto
      POSTGRES_PASSWORD: apisto
      POSTGRES_DB: apisto
    ports:
      - "5432:5432"
    volumes:
      - timescaledb_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U apisto"]
      interval: 5s
      timeout: 5s
      retries: 5

  # Mosquitto MQTT Broker
  mosquitto:
    image: eclipse-mosquitto:2
    container_name: apisto-mqtt
    restart: unless-stopped
    ports:
      - "1883:1883"
      - "9001:9001"   # WebSocket transport (optional, for browser MQTT clients)
    volumes:
      - ./mosquitto/mosquitto.conf:/mosquitto/config/mosquitto.conf
      - mosquitto_data:/mosquitto/data
      - mosquitto_log:/mosquitto/log
    healthcheck:
      test: ["CMD", "mosquitto_sub", "-t", "$$SYS/#", "-C", "1", "-i", "healthcheck", "-W", "3"]
      interval: 10s
      timeout: 5s
      retries: 3

  # Apisto API Server
  apisto:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: apisto-server
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      APISTO_PORT: 8080
      APISTO_DATABASE_URL: postgres://apisto:apisto@timescaledb:5432/apisto?sslmode=disable
      APISTO_MQTT_BROKER_URL: tcp://mosquitto:1883
      APISTO_MQTT_CLIENT_ID: apisto-server
      APISTO_RETENTION_DAYS: 30
      APISTO_HEARTBEAT_TIMEOUT: 60
      APISTO_LOG_LEVEL: info
      APISTO_CORS_ORIGINS: "*"
    depends_on:
      timescaledb:
        condition: service_healthy
      mosquitto:
        condition: service_healthy

volumes:
  timescaledb_data:
  mosquitto_data:
  mosquitto_log:
```

### mosquitto/mosquitto.conf

```
# Mosquitto configuration for Apisto
listener 1883
protocol mqtt

# WebSocket listener (optional — useful for browser-based MQTT clients in future)
listener 9001
protocol websockets

# Allow anonymous for v1 — auth handled at application layer
allow_anonymous true

# Persistence
persistence true
persistence_location /mosquitto/data/

# Logging
log_dest stdout
log_type error
log_type warning
log_type notice
log_type information

# Connection limits
max_connections -1
```

### Dockerfile (Apisto Go server only)

```dockerfile
# Build stage - Frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build stage - Backend
FROM golang:1.22-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o apisto ./cmd/apisto/

# Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/apisto .
EXPOSE 8080
CMD ["./apisto"]
```

---

## README.md

Write a clean, concise README with:

1. **Hero section:** One-liner description, a screenshot/GIF placeholder, and badges
2. **Quick Start:** Simple commands to get running:
   ```
   git clone https://github.com/your-org/apisto.git
   cd apisto
   docker compose up -d
   ```
   Then open `http://localhost:8080` in your browser.
3. **What is Apisto:** 2-3 sentences
4. **Features:** Bullet list (keep short)
5. **Arduino Quick Start:** Show the basic_telemetry.ino example
6. **HTTP API Alternative:** Show a curl example for HTTP telemetry ingestion
7. **Configuration:** Table of env vars
8. **Roadmap:** Brief list of what's coming (OTA, more SDKs, rules engine, fleet management)
9. **Contributing:** Standard section
10. **License:** MIT

---

## Implementation Order

Build in this exact sequence:

1. **Project scaffolding** — Go module, directory structure, Makefile, docker-compose.yml, mosquitto.conf, Dockerfile skeleton
2. **Config system** — env var loading with defaults (database URL, MQTT broker URL, ports)
3. **Docker Compose stack** — get TimescaleDB, Mosquitto, and Go server skeleton all booting with `docker compose up`
4. **PostgreSQL/TimescaleDB setup** — connection pool, golang-migrate runner, initial migration (all tables + hypertable + extensions)
5. **TimescaleDB policies** — second migration: retention policy, compression policy, continuous aggregates
6. **Models and database layer** — CRUD for projects, devices with Postgres
7. **MQTT client** — connect to Mosquitto using paho.mqtt.golang, subscribe to wildcard topics, log incoming messages
8. **MQTT telemetry handler** — parse incoming telemetry, validate device token, store in TimescaleDB hypertable, update device_keys
9. **Device status tracking** — track online/offline via MQTT message activity, heartbeat checker goroutine
10. **REST API** — projects CRUD, devices CRUD, telemetry query endpoints (with time_bucket aggregations), HTTP telemetry ingestion
11. **WebSocket hub** — real-time fan-out when telemetry arrives from MQTT
12. **Command system** — REST endpoint to send commands, MQTT publish to device topic, ack tracking
13. **Frontend shell** — Vite + React + Tailwind + React Router setup, API proxy config for dev
14. **Projects page** — list, create, delete projects
15. **Device list** — sidebar with devices, online/offline badges
16. **Auto-Dashboard** — widget rendering based on device_keys, real-time updates via WebSocket
17. **Control widgets** — toggle, slider, button UI → command API
18. **Quick Start page** — code snippets pre-filled with tokens
19. **Dashboard sharing** — share token generation, public routes
20. **Arduino SDK** — Apisto.h/cpp, example sketches
21. **Docker production build** — multi-stage Dockerfile, test full compose stack, test on ARM64
22. **README** — documentation

---

## Key Implementation Notes

### PostgreSQL Connection Pool
```go
// Use pgxpool for connection pooling
import "github.com/jackc/pgx/v5/pgxpool"

config, _ := pgxpool.ParseConfig(os.Getenv("APISTO_DATABASE_URL"))
config.MaxConns = 20
config.MinConns = 5
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute

pool, err := pgxpool.NewWithConfig(ctx, config)
```

### MQTT Client Setup
```go
// Use paho.mqtt.golang
import mqtt "github.com/eclipse/paho.mqtt.golang"

opts := mqtt.NewClientOptions()
opts.AddBroker(os.Getenv("APISTO_MQTT_BROKER_URL"))
opts.SetClientID(os.Getenv("APISTO_MQTT_CLIENT_ID"))
opts.SetAutoReconnect(true)
opts.SetMaxReconnectInterval(10 * time.Second)
opts.SetOnConnectHandler(func(c mqtt.Client) {
    // Re-subscribe to topics on reconnect
    c.Subscribe("apisto/+/telemetry", 1, telemetryHandler)
    c.Subscribe("apisto/+/status", 1, statusHandler)
    c.Subscribe("apisto/+/commands/ack", 1, ackHandler)
})

client := mqtt.NewClient(opts)
```

### MQTT Message Processing
Telemetry ingestion from MQTT should be fast. Use a buffered channel + worker goroutine pattern:
- MQTT message handler puts parsed telemetry into a buffered channel (non-blocking)
- Worker goroutine pool reads from channel and batch-inserts into TimescaleDB
- This prevents MQTT callback from blocking on DB writes
- Cache device token → device_id mappings in a sync.Map with TTL to avoid DB lookups on every message
- Use Postgres COPY protocol or batch INSERT for high-throughput telemetry ingestion

### Graceful Shutdown
Handle SIGINT/SIGTERM:
1. Disconnect MQTT client from Mosquitto (cleanly unsubscribes)
2. Close WebSocket connections
3. Flush telemetry buffer (drain the channel, write remaining data)
4. Close PostgreSQL connection pool
5. Exit

### Error Handling
- API returns consistent JSON errors: `{ "error": "message", "code": "ERROR_CODE" }`
- Invalid MQTT device tokens result in messages being silently dropped (logged at debug level)
- WebSocket disconnects are handled gracefully with auto-cleanup
- PostgreSQL connection failures trigger retry with backoff
- MQTT broker disconnection triggers automatic reconnect via paho client

### Testing
- Write integration tests for the core flow: create project → create device → send MQTT telemetry → query via API → verify data in TimescaleDB
- Test the auto-type detection for telemetry values
- Test MQTT token validation (valid token stores data, invalid token drops)
- Test WebSocket real-time delivery
- Test TimescaleDB aggregations (time_bucket queries, continuous aggregate queries)
- Use `testcontainers-go` to spin up TimescaleDB and Mosquitto for integration tests

---

## What is NOT in v1 (do not build these)

- No user authentication / login system (single-user, self-hosted)
- No OTA firmware updates
- No rules engine or alerting
- No data export
- No MicroPython / Python / Zephyr SDKs
- No device groups or fleet management
- No Grafana plugin
- No mobile app
- No edge compute
- No scheduled commands
- No multi-tenancy
- No rate limiting
- No TLS/SSL termination (users put nginx/caddy in front)

Keep the scope tight. Ship a complete, polished v1 rather than a sprawling incomplete v2.
