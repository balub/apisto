/**
 * Apisto — Full Example
 *
 * Combines telemetry (DHT22 sensor) with command handling (relay + LED).
 * Demonstrates all SDK features in a single sketch.
 *
 * Requirements:
 *   - PubSubClient library
 *   - ArduinoJson library
 *   - DHT sensor library (Adafruit)
 *
 * Hardware:
 *   - ESP32 or ESP8266
 *   - DHT22 on pin 4
 *   - Relay on pin 5
 *   - LED on pin 2
 */

#include <Apisto.h>
#include <DHT.h>

// ─── Configuration ────────────────────────────────────────
#define WIFI_SSID     "YourWiFiNetwork"
#define WIFI_PASSWORD "YourWiFiPassword"
#define DEVICE_TOKEN  "your_device_token_here"
#define SERVER_HOST   "192.168.1.100"

#define DHT_PIN   4
#define RELAY_PIN 5
#define LED_PIN   2

#define SEND_INTERVAL 10000  // ms between telemetry sends
// ──────────────────────────────────────────────────────────

DHT dht(DHT_PIN, DHT22);
Apisto device(DEVICE_TOKEN, SERVER_HOST);

bool relayState = false;

void setup() {
  Serial.begin(115200);
  dht.begin();
  pinMode(RELAY_PIN, OUTPUT);
  pinMode(LED_PIN, OUTPUT);

  device.begin(WIFI_SSID, WIFI_PASSWORD);

  // ─── Command handlers ──────────────────────────────────
  device.onCommand("relay", [](String payload) {
    relayState = (payload == "on");
    digitalWrite(RELAY_PIN, relayState ? HIGH : LOW);
    Serial.printf("[CMD] relay -> %s\n", relayState ? "on" : "off");
  });

  device.onCommand("led", [](String payload) {
    analogWrite(LED_PIN, constrain(payload.toInt(), 0, 255));
    Serial.printf("[CMD] led -> %s\n", payload.c_str());
  });

  device.onCommand("reboot", [](String payload) {
    Serial.println("[CMD] Rebooting...");
    delay(500);
    ESP.restart();
  });

  Serial.println("Apisto full example ready");
  Serial.printf("Server: %s:1883\n", SERVER_HOST);
  Serial.printf("Token:  %s\n", DEVICE_TOKEN);
}

void loop() {
  device.loop();

  static unsigned long lastSend = 0;
  if (millis() - lastSend > SEND_INTERVAL) {
    float temp = dht.readTemperature();
    float hum  = dht.readHumidity();

    device.beginBatch();

    if (!isnan(temp)) device.add("temperature", temp);
    if (!isnan(hum))  device.add("humidity", hum);

    device.add("relay_state", relayState);
    device.add("uptime_seconds", (int)(millis() / 1000));
    device.add("wifi_rssi", (int)WiFi.RSSI());
    device.add("free_heap", (int)ESP.getFreeHeap());

    device.endBatch();

    Serial.printf("Sent: temp=%.1f hum=%.1f relay=%s uptime=%ds\n",
      temp, hum, relayState ? "on" : "off", (int)(millis() / 1000));

    lastSend = millis();
  }
}
