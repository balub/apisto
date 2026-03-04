#include "Apisto.h"

Apisto* Apisto::_instance = nullptr;

Apisto::Apisto(const char* deviceToken, const char* serverHost, int mqttPort)
  : _token(deviceToken), _host(serverHost), _port(mqttPort),
    _ssid(nullptr), _password(nullptr),
    _mqtt(_wifiClient), _commandCount(0), _batching(false),
    _lastReconnectAttempt(0) {

  snprintf(_telemetryTopic, sizeof(_telemetryTopic), "apisto/%s/telemetry", _token);
  snprintf(_statusTopic, sizeof(_statusTopic), "apisto/%s/status", _token);
  snprintf(_commandTopic, sizeof(_commandTopic), "apisto/%s/commands", _token);

  _instance = this;
  _mqtt.setCallback(_mqttCallback);
  _mqtt.setBufferSize(1024);
}

void Apisto::begin(const char* wifiSSID, const char* wifiPassword) {
  _ssid = wifiSSID;
  _password = wifiPassword;
  _connectWifi();
  _connectMqtt();
}

void Apisto::loop() {
  if (WiFi.status() != WL_CONNECTED) {
    _connectWifi();
    return;
  }

  if (!_mqtt.connected()) {
    unsigned long now = millis();
    if (now - _lastReconnectAttempt > APISTO_RECONNECT_INTERVAL) {
      _lastReconnectAttempt = now;
      _connectMqtt();
    }
    return;
  }

  _mqtt.loop();
}

void Apisto::send(const char* key, float value) {
  StaticJsonDocument<256> doc;
  doc[key] = value;
  char buf[256];
  serializeJson(doc, buf);
  _publishTelemetry(buf);
}

void Apisto::send(const char* key, int value) {
  StaticJsonDocument<256> doc;
  doc[key] = value;
  char buf[256];
  serializeJson(doc, buf);
  _publishTelemetry(buf);
}

void Apisto::send(const char* key, bool value) {
  StaticJsonDocument<256> doc;
  doc[key] = value;
  char buf[256];
  serializeJson(doc, buf);
  _publishTelemetry(buf);
}

void Apisto::send(const char* key, const char* value) {
  StaticJsonDocument<256> doc;
  doc[key] = value;
  char buf[256];
  serializeJson(doc, buf);
  _publishTelemetry(buf);
}

void Apisto::beginBatch() {
  _batchDoc.clear();
  _batching = true;
}

void Apisto::add(const char* key, float value)       { if (_batching) _batchDoc[key] = value; }
void Apisto::add(const char* key, int value)         { if (_batching) _batchDoc[key] = value; }
void Apisto::add(const char* key, bool value)        { if (_batching) _batchDoc[key] = value; }
void Apisto::add(const char* key, const char* value) { if (_batching) _batchDoc[key] = value; }

void Apisto::endBatch() {
  if (!_batching) return;
  _batching = false;
  char buf[1024];
  serializeJson(_batchDoc, buf);
  _publishTelemetry(buf);
}

void Apisto::onCommand(const char* commandName, void (*callback)(String payload)) {
  if (_commandCount >= APISTO_MAX_COMMANDS) return;
  _commands[_commandCount++] = { commandName, callback };
}

bool Apisto::isConnected() {
  return WiFi.status() == WL_CONNECTED && _mqtt.connected();
}

void Apisto::_connectWifi() {
  if (WiFi.status() == WL_CONNECTED) return;

  Serial.print("[Apisto] Connecting to WiFi: ");
  Serial.println(_ssid);

  WiFi.begin(_ssid, _password);
  unsigned long start = millis();
  while (WiFi.status() != WL_CONNECTED && millis() - start < 15000) {
    delay(500);
    Serial.print(".");
  }

  if (WiFi.status() == WL_CONNECTED) {
    Serial.println("\n[Apisto] WiFi connected. IP: " + WiFi.localIP().toString());
  } else {
    Serial.println("\n[Apisto] WiFi connection failed");
  }
}

void Apisto::_connectMqtt() {
  if (!_mqtt.connected()) {
    _mqtt.setServer(_host, _port);

    char clientId[64];
    snprintf(clientId, sizeof(clientId), "apisto-%s", _token);

    Serial.print("[Apisto] Connecting to MQTT broker...");
    if (_mqtt.connect(clientId)) {
      Serial.println(" connected");
      _mqtt.subscribe(_commandTopic);
      // Publish online status
      _mqtt.publish(_statusTopic, "{\"status\":\"online\"}");
    } else {
      Serial.print(" failed, rc=");
      Serial.println(_mqtt.state());
    }
  }
}

void Apisto::_publishTelemetry(const char* payload) {
  if (!isConnected()) return;
  _mqtt.publish(_telemetryTopic, payload);
}

void Apisto::_onMqttMessage(char* topic, byte* payload, unsigned int length) {
  // Parse JSON command payload
  StaticJsonDocument<512> doc;
  DeserializationError err = deserializeJson(doc, payload, length);
  if (err) return;

  const char* command = doc["command"] | "";
  const char* cmdPayload = doc["payload"] | "";
  const char* cmdId = doc["id"] | "";

  // Find and call matching handler
  for (int i = 0; i < _commandCount; i++) {
    if (strcmp(_commands[i].name, command) == 0) {
      _commands[i].callback(String(cmdPayload));
      break;
    }
  }

  // Send acknowledgment
  if (strlen(cmdId) > 0) {
    char ackTopic[128];
    snprintf(ackTopic, sizeof(ackTopic), "apisto/%s/commands/ack", _token);
    _mqtt.publish(ackTopic, cmdId);
  }
}

void Apisto::_mqttCallback(char* topic, byte* payload, unsigned int length) {
  if (_instance) {
    _instance->_onMqttMessage(topic, payload, length);
  }
}
