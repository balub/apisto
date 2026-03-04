# Apisto — System Architecture

## Overview

Apisto is a single-binary Go server that orchestrates three subsystems: a REST/WebSocket API, an MQTT client, and a database layer. External infrastructure (TimescaleDB, Mosquitto) is managed via Docker Compose.

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Compose Stack                      │
│                                                             │
│  ┌──────────────┐   ┌──────────────┐   ┌────────────────┐  │
│  │  TimescaleDB │   │  Mosquitto   │   │  Apisto Server │  │
│  │  (Postgres)  │   │  MQTT Broker │   │  (Go binary)   │  │
│  │  :5432       │   │  :1883/:9001 │   │  :8080         │  │
│  └──────┬───────┘   └──────┬───────┘   └───────┬────────┘  │
│         │                  │                   │            │
└─────────┼──────────────────┼───────────────────┼────────────┘
          │                  │                   │
          └──────────────────┴───────────────────┘
                    Internal Docker network
```

## Apisto Server Internals

```
cmd/apisto/main.go
       │
       ▼
internal/server/server.go  (orchestrator)
       │
       ├── internal/config/config.go      (env vars + defaults)
       │
       ├── internal/database/postgres.go  (pgxpool + migrations)
       │
       ├── internal/mqtt/client.go        (paho MQTT client)
       │       │
       │       ├── Subscribe: apisto/+/telemetry
       │       ├── Subscribe: apisto/+/status
       │       └── Subscribe: apisto/+/commands/ack
       │
       ├── internal/api/router.go         (Fiber HTTP server)
       │       │
       │       ├── /api/v1/projects       (CRUD)
       │       ├── /api/v1/devices        (CRUD + keys)
       │       ├── /api/v1/*/telemetry    (ingest + query)
       │       ├── /api/v1/*/commands     (send + history)
       │       ├── /api/v1/*/ws           (WebSocket upgrade)
       │       └── /api/v1/public/*       (shared dashboards)
       │
       └── Background goroutines
               ├── Heartbeat checker (30s tick)
               └── Telemetry worker pool (buffered channel)
```

## Data Flow

### Device → Server (MQTT Telemetry)

```
ESP32/ESP8266
    │
    │ MQTT publish: apisto/{token}/telemetry
    │ Payload: {"temperature": 24.5, "humidity": 60}
    ▼
Mosquitto Broker
    │
    │ Forwards to all subscribers
    ▼
Apisto MQTT Client (paho)
    │
    ├── Extract token from topic
    ├── Validate token (sync.Map cache → DB fallback)
    ├── Push to buffered channel (non-blocking)
    │
    ▼
Telemetry Worker Pool
    │
    ├── Parse JSON payload
    ├── Detect value types (number/boolean/string)
    ├── Batch INSERT into telemetry hypertable
    ├── UPSERT device_keys (auto-discovery)
    ├── UPDATE device.last_seen_at
    │
    ▼
WebSocket Hub (fan-out)
    │
    └── Push to all connected browser clients for that device_id
```

### Server → Device (Commands)

```
Browser Dashboard
    │
    │ POST /api/v1/devices/:id/commands
    │ Body: {"command": "relay", "payload": "on"}
    ▼
Apisto API Handler
    │
    ├── INSERT into commands table (status: pending)
    ├── Look up device token
    │
    ▼
Apisto MQTT Client
    │
    │ MQTT publish: apisto/{token}/commands
    │ Payload: {"id": "cmd_xyz", "command": "relay", "payload": "on"}
    ▼
Mosquitto Broker
    │
    ▼
ESP32/ESP8266
    │
    │ MQTT publish: apisto/{token}/commands/ack
    │ Payload: {"id": "cmd_xyz", "status": "acknowledged"}
    ▼
Apisto MQTT Client
    │
    └── UPDATE commands SET status='acknowledged', acked_at=NOW()
```

### Browser → Live Dashboard (WebSocket)

```
Browser
    │
    │ GET /api/v1/devices/:id/ws (WebSocket upgrade)
    ▼
Apisto WebSocket Handler
    │
    └── Register connection in hub: device_id → conn

[When MQTT telemetry arrives for that device_id]
    │
    ▼
WebSocket Hub
    │
    └── Fan-out JSON message to all registered conns for device_id
```

## MQTT Topic Structure

| Topic | Direction | Purpose |
|-------|-----------|---------|
| `apisto/{token}/telemetry` | Device → Server | Sensor data |
| `apisto/{token}/status` | Device → Server | Heartbeat/online status |
| `apisto/{token}/commands` | Server → Device | Command dispatch |
| `apisto/{token}/commands/ack` | Device → Server | Command acknowledgment |

## Database Architecture

Single PostgreSQL instance with TimescaleDB extension:

```
PostgreSQL 16 + TimescaleDB
│
├── projects          (relational)
├── devices           (relational)
├── device_keys       (relational — auto-discovered telemetry keys)
├── commands          (relational — command log)
├── dashboard_shares  (relational — public share tokens)
│
└── telemetry         (TimescaleDB hypertable — time-series)
        │
        ├── Partitioned by time (1-day chunks)
        ├── Compressed after 7 days
        ├── Dropped after 30 days (retention policy)
        │
        ├── telemetry_hourly  (continuous aggregate)
        └── telemetry_daily   (continuous aggregate)
```

## Scalability Notes

- Designed for single-node deployment (hobbyist/maker use case)
- Target: Raspberry Pi 4 (ARM64) with 1–4 devices, 1–100 data points/minute
- TimescaleDB handles time-series scale up to millions of rows with partitioning
- pgxpool handles up to 20 concurrent DB connections
- MQTT worker pool prevents DB writes from blocking message ingestion
- For higher scale: add read replicas, tune chunk intervals, use continuous aggregates for dashboards
