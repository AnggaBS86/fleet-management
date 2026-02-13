package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"leet-management/internal/config"
	"leet-management/internal/models"
)

func main() {
	cfg := config.Load()

	vehicleID := os.Getenv("VEHICLE_ID")
	if vehicleID == "" {
		vehicleID = "B1234XYZ"
	}

	baseLat := getEnvFloat("BASE_LAT", -6.2088)
	baseLon := getEnvFloat("BASE_LON", 106.8456)

	opts := paho.NewClientOptions().AddBroker(cfg.MQTTBroker).SetClientID("fleet-publisher-" + vehicleID)
	client := paho.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt connect failed: %v", token.Error())
	}
	defer client.Disconnect(250)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	topic := "/fleet/vehicle/" + vehicleID + "/location"
	log.Printf("publishing to %s", topic)

	for {
		latOffset := (rng.Float64() - 0.5) * 0.002
		lonOffset := (rng.Float64() - 0.5) * 0.002

		msg := models.LocationMessage{
			VehicleID: vehicleID,
			Latitude:  round(baseLat+latOffset, 6),
			Longitude: round(baseLon+lonOffset, 6),
			Timestamp: time.Now().Unix(),
		}

		payload, _ := json.Marshal(msg)
		if token := client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			log.Printf("publish failed: %v", token.Error())
		}

		time.Sleep(2 * time.Second)
	}
}

func round(value float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(value*pow) / pow
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
