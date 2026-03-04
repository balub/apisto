# ⬡ Apisto

**Self-hosted, lightweight IoT backend for makers.** Connect your ESP32/ESP8266, stream telemetry, visualize in real-time, and control devices remotely — from `docker compose up` to a live dashboard in under 15 minutes.

![License](https://img.shields.io/badge/license-MIT-blue)
![Go](https://img.shields.io/badge/go-1.22-00ADD8)
![Docker](https://img.shields.io/badge/docker-compose-2496ED)

---

## Quick Start

```bash
git clone https://github.com/balub/apisto.git
cd apisto
docker compose up -d
```

Open **http://localhost:8080** and create your first project.

---

## What is Apisto?

Apisto is a single Go binary that acts as a Firebase for IoT — purpose-built for hobbyists and makers. It speaks MQTT natively (via Mosquitto), stores time-series data in TimescaleDB, auto-generates dashboards from your device's telemetry, and lets you send commands back to your hardware from the browser.

**One server. One `docker compose up`. Zero configuration.**

---

## Features

- **Auto-Dashboard** — widgets appear automatically for each telemetry key (numeric cards, charts, boolean indicators, event logs)
- **Real-time updates** — WebSocket push from device to browser with no polling
- **MQTT + HTTP** — connect via MQTT (Arduino SDK) or plain HTTP (curl, Python)
- **TimescaleDB** — time-series storage with automatic compression and retention policies
- **Command & Control** — send commands to devices from the UI, with acknowledgment tracking
- **Dashboard Sharing** — share a public read-only view of any device (no login required)
- **Arduino SDK** — single `.h`/`.cpp` library for ESP32/ESP8266 with auto-reconnect

---

## Arduino Quick Start

Install these Arduino libraries: **PubSubClient** + **ArduinoJson**. Copy `sdk/arduino/Apisto.h` and `Apisto.cpp` into your project.

```cpp
#include <Apisto.h>
#include <DHT.h>

#define DEVICE_TOKEN "your_device_token_here"
#define SERVER_HOST  "192.168.1.100"  // your server IP

DHT dht(4, DHT22);
Apisto device(DEVICE_TOKEN, SERVER_HOST);

void setup() {
  dht.begin();
  device.begin("YourWiFi", "YourPassword");

  // Optional: handle commands from dashboard
  device.onCommand("relay", [](String payload) {
    digitalWrite(5, payload == "on" ? HIGH : LOW);
  });
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

---

## HTTP API (no Arduino required)

```bash
# Create a project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "My Home"}'

# Add a device (save the token!)
curl -X POST http://localhost:8080/api/v1/projects/PROJECT_ID/devices \
  -H "Content-Type: application/json" \
  -d '{"name": "Sensor 1"}'

# Send telemetry
curl -X POST http://localhost:8080/api/v1/devices/YOUR_TOKEN/telemetry \
  -H "Content-Type: application/json" \
  -d '{"temperature": 24.5, "humidity": 60, "relay_on": true}'

# Query telemetry (last hour, 1-minute averages)
curl "http://localhost:8080/api/v1/devices/DEVICE_ID/telemetry?key=temperature&aggregate=avg&interval=1m"
```

---

## Configuration

All settings via environment variables (see `.env.example`):

| Variable | Default | Description |
|----------|---------|-------------|
| `APISTO_PORT` | `8080` | HTTP/WebSocket port |
| `APISTO_DATABASE_URL` | `postgres://apisto:apisto@timescaledb:5432/apisto` | PostgreSQL connection string |
| `APISTO_MQTT_BROKER_URL` | `tcp://mosquitto:1883` | Mosquitto broker URL |
| `APISTO_MQTT_CLIENT_ID` | `apisto-server` | MQTT client ID |
| `APISTO_RETENTION_DAYS` | `30` | Telemetry retention in days |
| `APISTO_HEARTBEAT_TIMEOUT` | `60` | Seconds before device marked offline |
| `APISTO_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `APISTO_CORS_ORIGINS` | `*` | Allowed CORS origins |

---

## Architecture

```
Browser ──WebSocket──► Apisto Server ◄──MQTT──► Mosquitto ◄──MQTT──► ESP32
                            │
                            └──SQL──► TimescaleDB (PostgreSQL)
```

Single Go binary. Three Docker services. No external dependencies.

See [docs/architecture.md](docs/architecture.md) for the full data flow diagram.

---

## Documentation

- [Quick Start Guide](docs/quickstart.md) — step-by-step: docker up → first device → live dashboard
- [API Reference](docs/api.md) — all REST endpoints with examples
- [Architecture](docs/architecture.md) — system design and data flow
- [Full Spec](docs/spec.md) — complete build specification

---

## Roadmap

- [ ] OTA firmware updates via MQTT
- [ ] Rules engine (trigger actions on threshold)
- [ ] Data export (CSV, JSON)
- [ ] MicroPython SDK
- [ ] Grafana datasource plugin
- [ ] Fleet management (device groups, bulk commands)
- [ ] TLS/SSL termination built-in
- [ ] Scheduled commands

---

## Development

```bash
# Start dependencies
docker compose up timescaledb mosquitto -d

# Run backend (requires Go 1.22+)
go run ./cmd/apisto/

# Run frontend dev server (hot reload)
cd web && npm install && npm run dev
```

Backend API runs on `:8080`, Vite dev server on `:5173` with proxy.

---

## Contributing

1. Fork and create a feature branch
2. Make your changes with tests
3. Open a pull request — please describe what and why

Issues and feature requests welcome at [GitHub Issues](https://github.com/balub/apisto/issues).

---

## License

MIT

---

*Built with Go, React, TimescaleDB, and Mosquitto. Designed for Raspberry Pi and beyond.*
