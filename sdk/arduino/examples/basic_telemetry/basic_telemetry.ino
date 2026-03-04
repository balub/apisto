/**
 * Apisto — Basic Telemetry Example
 *
 * Sends temperature and humidity from a DHT22 sensor to Apisto every 5 seconds.
 *
 * Requirements:
 *   - PubSubClient library (by Nick O'Leary)
 *   - ArduinoJson library (by Benoit Blanchon)
 *   - DHT sensor library (by Adafruit)
 *
 * Hardware:
 *   - ESP32 or ESP8266
 *   - DHT22 sensor connected to pin 4
 */

#include <Apisto.h>
#include <DHT.h>

// ─── Configuration ────────────────────────────────────────
#define WIFI_SSID     "YourWiFiNetwork"
#define WIFI_PASSWORD "YourWiFiPassword"
#define DEVICE_TOKEN  "your_device_token_here"
#define SERVER_HOST   "192.168.1.100"   // Your Apisto server IP

#define DHT_PIN  4
#define DHT_TYPE DHT22
// ──────────────────────────────────────────────────────────

DHT dht(DHT_PIN, DHT_TYPE);
Apisto device(DEVICE_TOKEN, SERVER_HOST);

void setup() {
  Serial.begin(115200);
  dht.begin();
  device.begin(WIFI_SSID, WIFI_PASSWORD);
  Serial.println("Apisto basic telemetry example started");
}

void loop() {
  device.loop();

  static unsigned long lastSend = 0;
  if (millis() - lastSend > 5000) {
    float temp = dht.readTemperature();
    float hum  = dht.readHumidity();

    if (!isnan(temp) && !isnan(hum)) {
      // Send as a batch (single MQTT message)
      device.beginBatch();
      device.add("temperature", temp);
      device.add("humidity", hum);
      device.endBatch();

      Serial.printf("Sent: temp=%.1f°C hum=%.1f%%\n", temp, hum);
    } else {
      Serial.println("DHT read failed");
    }

    lastSend = millis();
  }
}
