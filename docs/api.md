# Apisto REST API Reference

Base URL: `http://localhost:8080/api/v1`

All responses are JSON. Errors use the format: `{"error": "message", "code": "ERROR_CODE"}`

---

## Projects

### Create Project
```
POST /api/v1/projects
```
**Body:**
```json
{
  "name": "My Smart Home",
  "description": "Home automation project"
}
```
**Response:**
```json
{
  "id": "a1b2c3d4",
  "name": "My Smart Home",
  "description": "Home automation project",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

### List Projects
```
GET /api/v1/projects
```
**Response:**
```json
[
  {
    "id": "a1b2c3d4",
    "name": "My Smart Home",
    "description": "Home automation project",
    "device_count": 3,
    "created_at": "2024-01-15T10:00:00Z"
  }
]
```

### Get Project
```
GET /api/v1/projects/:id
```

### Update Project
```
PUT /api/v1/projects/:id
```
**Body:** Same fields as create (all optional).

### Delete Project
```
DELETE /api/v1/projects/:id
```
Cascades: deletes all devices and their telemetry.

---

## Devices

### Create Device
```
POST /api/v1/projects/:id/devices
```
**Body:**
```json
{
  "name": "Living Room Sensor",
  "description": "Temperature and humidity monitor"
}
```
**Response:**
```json
{
  "id": "e5f6a7b8",
  "project_id": "a1b2c3d4",
  "token": "abc123def456abc123def456abc123de",
  "name": "Living Room Sensor",
  "description": "Temperature and humidity monitor",
  "is_online": false,
  "created_at": "2024-01-15T10:05:00Z"
}
```
The `token` is the MQTT device token — store it securely, it's shown only once.

### List Devices in Project
```
GET /api/v1/projects/:id/devices
```

### Get Device
```
GET /api/v1/devices/:id
```
**Response includes:** name, status, last_seen_at, ip_address, firmware_version, device_keys count.

### Update Device
```
PUT /api/v1/devices/:id
```
**Body:**
```json
{
  "name": "Updated Name",
  "description": "New description",
  "firmware_version": "1.2.3"
}
```

### Delete Device
```
DELETE /api/v1/devices/:id
```

### Get Device Keys
```
GET /api/v1/devices/:id/keys
```
Returns auto-discovered telemetry keys for this device.

**Response:**
```json
[
  {
    "key": "temperature",
    "value_type": "number",
    "widget_type": "auto",
    "display_name": "Temperature",
    "unit": "°C",
    "first_seen_at": "2024-01-15T10:10:00Z",
    "last_seen_at": "2024-01-15T12:00:00Z"
  },
  {
    "key": "relay_on",
    "value_type": "boolean",
    "widget_type": "auto",
    "display_name": "",
    "unit": "",
    "first_seen_at": "2024-01-15T10:10:00Z",
    "last_seen_at": "2024-01-15T12:00:00Z"
  }
]
```

---

## Telemetry

### HTTP Ingest (alternative to MQTT)
```
POST /api/v1/devices/:token/telemetry
```
Use the device **token** (not ID) in the URL.

**Body:**
```json
{
  "temperature": 24.5,
  "humidity": 60,
  "relay_on": true,
  "status_msg": "all good"
}
```

**curl example:**
```bash
curl -X POST http://localhost:8080/api/v1/devices/YOUR_TOKEN/telemetry \
  -H "Content-Type: application/json" \
  -d '{"temperature": 24.5, "humidity": 60}'
```

### Query Telemetry
```
GET /api/v1/devices/:id/telemetry
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `key` | string | (all keys) | Filter by telemetry key |
| `from` | ISO8601 | 1 hour ago | Start time |
| `to` | ISO8601 | now | End time |
| `limit` | int | 1000 | Max rows returned |
| `order` | `asc`/`desc` | `asc` | Sort order |
| `aggregate` | `avg`/`min`/`max`/`count`/`last` | (none) | Aggregation function |
| `interval` | `1m`/`5m`/`15m`/`1h`/`1d` | `1h` | Aggregation window (requires `aggregate`) |

**Example — last 24h of temperature, hourly averages:**
```
GET /api/v1/devices/e5f6a7b8/telemetry?key=temperature&from=2024-01-14T10:00:00Z&aggregate=avg&interval=1h
```

**Response:**
```json
{
  "device_id": "e5f6a7b8",
  "key": "temperature",
  "aggregate": "avg",
  "interval": "1h",
  "data": [
    {
      "time": "2024-01-14T10:00:00Z",
      "value": 23.4,
      "min": 22.1,
      "max": 24.8,
      "count": 720
    },
    {
      "time": "2024-01-14T11:00:00Z",
      "value": 24.1,
      "min": 23.5,
      "max": 25.2,
      "count": 718
    }
  ]
}
```

### Get Latest Values
```
GET /api/v1/devices/:id/telemetry/latest
```
Returns the most recent value for each discovered key.

**Response:**
```json
{
  "device_id": "e5f6a7b8",
  "data": {
    "temperature": {
      "value": 24.5,
      "value_type": "number",
      "time": "2024-01-15T12:00:00Z"
    },
    "relay_on": {
      "value": true,
      "value_type": "boolean",
      "time": "2024-01-15T12:00:00Z"
    },
    "status_msg": {
      "value": "all good",
      "value_type": "string",
      "time": "2024-01-15T11:55:00Z"
    }
  }
}
```

---

## Commands

### Send Command to Device
```
POST /api/v1/devices/:id/commands
```
**Body:**
```json
{
  "command": "relay",
  "payload": "on"
}
```
**Response:**
```json
{
  "id": "cmd_xyz789",
  "device_id": "e5f6a7b8",
  "command": "relay",
  "payload": "on",
  "status": "pending",
  "created_at": "2024-01-15T12:05:00Z"
}
```

The command is immediately published to the device via MQTT. Status transitions:
- `pending` → command created
- `sent` → published to MQTT broker
- `acknowledged` → device confirmed receipt
- `failed` → delivery failed
- `expired` → device never acknowledged

### List Command History
```
GET /api/v1/devices/:id/commands
```
**Query params:** `limit=50`, `status=acknowledged`

---

## WebSocket

### Device Real-time Stream
```
GET /api/v1/devices/:id/ws
```
Upgrade to WebSocket. Receives real-time telemetry, status, and command ack events.

**Message types:**

```json
{
  "type": "telemetry",
  "device_id": "e5f6a7b8",
  "data": { "temperature": 24.5, "humidity": 60 },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

```json
{
  "type": "status",
  "device_id": "e5f6a7b8",
  "is_online": true,
  "last_seen_at": "2024-01-15T10:30:00Z"
}
```

```json
{
  "type": "command_ack",
  "command_id": "cmd_xyz789",
  "status": "acknowledged"
}
```

**JavaScript example:**
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/devices/e5f6a7b8/ws');
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'telemetry') {
    console.log('New data:', msg.data);
  }
};
```

---

## Dashboard Shares

### Create Share Link
```
POST /api/v1/devices/:id/share
```
**Body (optional):**
```json
{
  "password": "optional_password"
}
```
**Response:**
```json
{
  "id": "share_abc",
  "share_token": "public_share_token_here",
  "url": "http://localhost:8080/public/public_share_token_here",
  "created_at": "2024-01-15T12:00:00Z"
}
```

### Revoke Share
```
DELETE /api/v1/shares/:share_token
```

### Public Dashboard Data (no auth)
```
GET /api/v1/public/:share_token
```
Returns device data (latest telemetry, device name) for public display.

### Public WebSocket (no auth)
```
GET /api/v1/public/:share_token/ws
```
Real-time stream for shared dashboard.

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NOT_FOUND` | 404 | Resource not found |
| `INVALID_INPUT` | 400 | Request validation failed |
| `INVALID_TOKEN` | 401 | Invalid device token |
| `INTERNAL_ERROR` | 500 | Server error |
