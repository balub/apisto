package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	MQTTBrokerURL     string
	MQTTClientID      string
	RetentionDays     int
	HeartbeatTimeout  int
	LogLevel          string
	CORSOrigins       string
}

func Load() *Config {
	return &Config{
		Port:             getEnv("APISTO_PORT", "8080"),
		DatabaseURL:      getEnv("APISTO_DATABASE_URL", "postgres://apisto:apisto@localhost:5432/apisto?sslmode=disable"),
		MQTTBrokerURL:    getEnv("APISTO_MQTT_BROKER_URL", "tcp://localhost:1883"),
		MQTTClientID:     getEnv("APISTO_MQTT_CLIENT_ID", "apisto-server"),
		RetentionDays:    getEnvInt("APISTO_RETENTION_DAYS", 30),
		HeartbeatTimeout: getEnvInt("APISTO_HEARTBEAT_TIMEOUT", 60),
		LogLevel:         getEnv("APISTO_LOG_LEVEL", "info"),
		CORSOrigins:      getEnv("APISTO_CORS_ORIGINS", "*"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}
