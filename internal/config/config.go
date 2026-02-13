package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort              string
	PostgresURL           string
	MQTTBroker            string
	MQTTClientID          string
	MQTTUsername          string
	MQTTPassword          string
	RabbitURL             string
	GeofenceLat           float64
	GeofenceLon           float64
	GeofenceRadiusMeters  float64
	GeofenceEventExchange string
	GeofenceQueue         string
	GeofenceRoutingKey    string
}

func Load() Config {
	return Config{
		HTTPPort:              getEnv("HTTP_PORT", "8080"),
		PostgresURL:           getEnv("POSTGRES_URL", "postgres://fleet:fleet@localhost:5432/fleet?sslmode=disable"),
		MQTTBroker:            getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTClientID:          getEnv("MQTT_CLIENT_ID", "fleet-api"),
		MQTTUsername:          getEnv("MQTT_USERNAME", ""),
		MQTTPassword:          getEnv("MQTT_PASSWORD", ""),
		RabbitURL:             getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
		GeofenceLat:           getEnvFloat("GEOFENCE_LAT", -6.2088),
		GeofenceLon:           getEnvFloat("GEOFENCE_LON", 106.8456),
		GeofenceRadiusMeters:  getEnvFloat("GEOFENCE_RADIUS_METERS", 50),
		GeofenceEventExchange: getEnv("RABBIT_EXCHANGE", "fleet.events"),
		GeofenceQueue:         getEnv("RABBIT_QUEUE", "geofence_alerts"),
		GeofenceRoutingKey:    getEnv("RABBIT_ROUTING_KEY", "geofence_alerts"),
	}
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getEnvFloat(key string, def float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return def
	}
	return parsed
}
