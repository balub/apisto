/**
 * Apisto — Command & Control Example
 *
 * Demonstrates receiving commands from the Apisto dashboard
 * to control a relay and LED brightness.
 *
 * Requirements:
 *   - PubSubClient library
 *   - ArduinoJson library
 *
 * Hardware:
 *   - ESP32 or ESP8266
 *   - Relay on pin 5
 *   - LED on pin 2 (built-in on most ESP boards)
 */

#include <Apisto.h>

// ─── Configuration ────────────────────────────────────────
#define WIFI_SSID     "YourWiFiNetwork"
#define WIFI_PASSWORD "YourWiFiPassword"
#define DEVICE_TOKEN  "your_device_token_here"
#define SERVER_HOST   "192.168.1.100"

#define RELAY_PIN 5
#define LED_PIN   2
// ──────────────────────────────────────────────────────────

Apisto device(DEVICE_TOKEN, SERVER_HOST);

void setup() {
  Serial.begin(115200);
  pinMode(RELAY_PIN, OUTPUT);
  pinMode(LED_PIN, OUTPUT);
  digitalWrite(RELAY_PIN, LOW);
  digitalWrite(LED_PIN, LOW);

  device.begin(WIFI_SSID, WIFI_PASSWORD);

  // Register command handlers
  device.onCommand("relay", [](String payload) {
    bool on = (payload == "on" || payload == "1" || payload == "true");
    digitalWrite(RELAY_PIN, on ? HIGH : LOW);
    Serial.println("Relay: " + payload);
  });

  device.onCommand("led", [](String payload) {
    int brightness = payload.toInt();
    brightness = constrain(brightness, 0, 255);
    analogWrite(LED_PIN, brightness);
    Serial.println("LED brightness: " + String(brightness));
  });

  device.onCommand("blink", [](String payload) {
    int times = max(1, payload.toInt());
    for (int i = 0; i < times; i++) {
      digitalWrite(LED_PIN, HIGH);
      delay(200);
      digitalWrite(LED_PIN, LOW);
      delay(200);
    }
  });

  Serial.println("Apisto command example started. Waiting for commands...");
}

void loop() {
  device.loop();

  // Report state every 5 seconds
  static unsigned long lastSend = 0;
  if (millis() - lastSend > 5000) {
    device.beginBatch();
    device.add("relay_state", digitalRead(RELAY_PIN) == HIGH);
    device.add("uptime_seconds", (int)(millis() / 1000));
    device.endBatch();

    lastSend = millis();
  }
}
