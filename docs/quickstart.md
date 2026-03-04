# Apisto Quick Start

Get from zero to a live IoT dashboard in under 15 minutes.

---

## Prerequisites

- Docker and Docker Compose installed
- An ESP32 or ESP8266 board (or use curl to simulate a device)
- Arduino IDE with PubSubClient and ArduinoJson libraries

---

## Step 1: Start the Stack

```bash
git clone https://github.com/balub/apisto.git
cd apisto
docker compose up -d
```

Wait ~10 seconds for TimescaleDB to initialize. Check health:

```bash
docker compose ps
```

All three services should show `healthy` or `running`.

Open http://localhost:8080 — you should see the Apisto dashboard.

---

## Step 2: Create a Project

In the web UI, click **"New Project"** and give it a name (e.g., "My Home").

Or via curl:
```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "My Home", "description": "Home automation"}'
```

Note the returned `id` — this is your project ID.

---

## Step 3: Add a Device

In the web UI, open your project and click **"Add Device"**.

Or via curl (replace `PROJECT_ID`):
```bash
curl -X POST http://localhost:8080/api/v1/projects/PROJECT_ID/devices \
  -H "Content-Type: application/json" \
  -d '{"name": "Living Room Sensor"}'
```

**Important:** Copy the `token` from the response. You'll need it for your device. It's only shown once.

---

## Step 4: Connect Your Device

### Option A: Arduino (ESP32/ESP8266)

Install these Arduino libraries (via Library Manager):
- `PubSubClient` by Nick O'Leary
- `ArduinoJson` by Benoit Blanchon

Download [Apisto.h and Apisto.cpp](../sdk/arduino/) and add to your Arduino libraries folder.

Upload this sketch (replace the placeholders):

```cpp
#include <Apisto.h>
#include <DHT.h>

#define WIFI_SSID     "YourWiFiNetwork"
#define WIFI_PASSWORD "YourWiFiPassword"
#define DEVICE_TOKEN  "YOUR_DEVICE_TOKEN_HERE"
#define SERVER_HOST   "192.168.1.100"  // Your server's IP address

DHT dht(4, DHT22);
Apisto device(DEVICE_TOKEN, SERVER_HOST);

void setup() {
  Serial.begin(115200);
  dht.begin();
  device.begin(WIFI_SSID, WIFI_PASSWORD);
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

### Option B: curl (test without hardware)

Send telemetry via HTTP (replace `YOUR_TOKEN`):
```bash
curl -X POST http://localhost:8080/api/v1/devices/YOUR_TOKEN/telemetry \
  -H "Content-Type: application/json" \
  -d '{"temperature": 24.5, "humidity": 60, "relay_on": true}'
```

### Option C: MQTT (mosquitto_pub)

```bash
mosquitto_pub -h localhost -p 1883 \
  -t "apisto/YOUR_TOKEN/telemetry" \
  -m '{"temperature": 24.5, "humidity": 60}'
```

---

## Step 5: See Your Data

Open http://localhost:8080 and navigate to your project. You should see:

- Your device listed in the sidebar with a **green online indicator**
- Auto-generated widgets for each telemetry key:
  - Temperature → numeric card with sparkline
  - Humidity → numeric card with sparkline
  - relay_on → green/red boolean indicator
- Real-time updates via WebSocket (no page refresh needed)

---

## Step 6: Send Commands (Optional)

From the device view in the UI, use the Controls section to send commands.

Or via curl:
```bash
curl -X POST http://localhost:8080/api/v1/devices/DEVICE_ID/commands \
  -H "Content-Type: application/json" \
  -d '{"command": "relay", "payload": "on"}'
```

Your Arduino sketch can receive commands:
```cpp
device.onCommand("relay", [](String payload) {
  digitalWrite(RELAY_PIN, payload == "on" ? HIGH : LOW);
});
```

---

## Step 7: Share Your Dashboard (Optional)

Create a public share link (no login required to view):
```bash
curl -X POST http://localhost:8080/api/v1/devices/DEVICE_ID/share
```

Share the returned URL with anyone — they can view live data without credentials.

---

## Troubleshooting

**Device not showing as online:**
- Check that your device token matches exactly
- Verify the server IP is reachable from your device
- Check MQTT port 1883 is not firewalled

**No telemetry data:**
- Try the curl method first to verify the API works
- Check `docker compose logs apisto` for errors
- Ensure your JSON payload is valid

**TimescaleDB startup slow:**
- First startup can take 30-60 seconds for initialization
- Run `docker compose logs timescaledb` to monitor

**Finding your server's IP:**
```bash
# Linux/Mac
hostname -I | awk '{print $1}'

# Or check your router's admin page
```
