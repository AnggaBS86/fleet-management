package mqtt

import (
	"context"
	"encoding/json"
	"log"

	paho "github.com/eclipse/paho.mqtt.golang"

	"leet-management/internal/db"
	"leet-management/internal/geofence"
	"leet-management/internal/models"
	"leet-management/internal/queue"
)

type Subscriber struct {
	client      paho.Client
	store       *db.Store
	publisher   *queue.Publisher
	geofenceLat float64
	geofenceLon float64
	radiusM     float64
}

func NewSubscriber(broker, clientID, username, password string, store *db.Store, publisher *queue.Publisher, geofenceLat, geofenceLon, radiusM float64) (*Subscriber, error) {
	opts := paho.NewClientOptions().AddBroker(broker).SetClientID(clientID)
	if username != "" {
		opts.SetUsername(username)
		opts.SetPassword(password)
	}

	client := paho.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &Subscriber{
		client:      client,
		store:       store,
		publisher:   publisher,
		geofenceLat: geofenceLat,
		geofenceLon: geofenceLon,
		radiusM:     radiusM,
	}, nil
}

func (s *Subscriber) Close() {
	if s.client != nil {
		s.client.Disconnect(250)
	}
}

func (s *Subscriber) Start(ctx context.Context, topic string) error {
	handler := func(_ paho.Client, msg paho.Message) {
		var payload models.LocationMessage
		if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
			log.Printf("mqtt: invalid json: %v", err)
			return
		}
		if !payload.Valid() {
			log.Printf("mqtt: invalid payload: %+v", payload)
			return
		}

		ctxTimeout, cancel := db.WithTimeout(ctx)
		defer cancel()
		if err := s.store.InsertLocation(ctxTimeout, payload.VehicleID, payload.Latitude, payload.Longitude, payload.Timestamp); err != nil {
			log.Printf("db insert failed: %v", err)
			return
		}

		// get distance, then if distacne is 50 meter --> publish to RabbitMQ
		distance := geofence.DistanceMeters(s.geofenceLat, s.geofenceLon, payload.Latitude, payload.Longitude)
		if distance <= s.radiusM {
			event := queue.GeofenceEvent{
				VehicleID: payload.VehicleID,
				Event:     "geofence_entry",
				Timestamp: payload.Timestamp,
			}
			event.Location.Latitude = payload.Latitude
			event.Location.Longitude = payload.Longitude

			if err := s.publisher.PublishToRabbitMq(event); err != nil {
				log.Printf("publish geofence event failed: %v", err)
			}
		}
	}

	if token := s.client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
