package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"

	"leet-management/internal/config"
)

func main() {
	cfg := config.Load()

	conn, err := amqp.Dial(cfg.RabbitURL)
	if err != nil {
		log.Fatalf("rabbit connect failed: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("rabbit channel failed: %v", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(cfg.GeofenceEventExchange, "direct", true, false, false, false, nil); err != nil {
		log.Fatalf("exchange declare failed: %v", err)
	}

	queue, err := ch.QueueDeclare(cfg.GeofenceQueue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue declare failed: %v", err)
	}

	if err := ch.QueueBind(queue.Name, cfg.GeofenceRoutingKey, cfg.GeofenceEventExchange, false, nil); err != nil {
		log.Fatalf("queue bind failed: %v", err)
	}

	msgs, err := ch.Consume(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue consume failed: %v", err)
	}

	log.Printf("worker listening on queue %s", queue.Name)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		for msg := range msgs {
			log.Printf("geofence event: %s", string(msg.Body))
		}
	}()

	<-stop
}
