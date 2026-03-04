#pragma once

#include <Arduino.h>
#include <ArduinoJson.h>
#include <PubSubClient.h>

#ifdef ESP32
#include <WiFi.h>
#elif defined(ESP8266)
#include <ESP8266WiFi.h>
#endif

#define APISTO_MQTT_PORT 1883
#define APISTO_MAX_COMMANDS 10
#define APISTO_RECONNECT_INTERVAL 5000

struct ApistoCommand {
  const char* name;
  void (*callback)(String payload);
};

class Apisto {
public:
  Apisto(const char* deviceToken, const char* serverHost, int mqttPort = APISTO_MQTT_PORT);

  // Call in setup()
  void begin(const char* wifiSSID, const char* wifiPassword);

  // Call in loop() — handles reconnects and incoming commands
  void loop();

  // Send a single telemetry value
  void send(const char* key, float value);
  void send(const char* key, int value);
  void send(const char* key, bool value);
  void send(const char* key, const char* value);

  // Batch multiple values into a single MQTT message
  void beginBatch();
  void add(const char* key, float value);
  void add(const char* key, int value);
  void add(const char* key, bool value);
  void add(const char* key, const char* value);
  void endBatch();

  // Register a command handler
  void onCommand(const char* commandName, void (*callback)(String payload));

  // Status
  bool isConnected();

private:
  const char* _token;
  const char* _host;
  int _port;
  const char* _ssid;
  const char* _password;

  WiFiClient _wifiClient;
  PubSubClient _mqtt;

  char _telemetryTopic[128];
  char _statusTopic[128];
  char _commandTopic[128];

  ApistoCommand _commands[APISTO_MAX_COMMANDS];
  int _commandCount;

  StaticJsonDocument<1024> _batchDoc;
  bool _batching;

  unsigned long _lastReconnectAttempt;

  void _connectWifi();
  void _connectMqtt();
  void _publishTelemetry(const char* payload);
  void _onMqttMessage(char* topic, byte* payload, unsigned int length);
  static void _mqttCallback(char* topic, byte* payload, unsigned int length);
  static Apisto* _instance;
};
