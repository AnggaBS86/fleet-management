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

	// ExchangeDeclare declares an exchange on the server. If the exchange does not already exist, the server will create it.
	// If the exchange exists, the server verifies that it is of the provided type, durability and auto-delete flags.
	// Errors returned from this method will close the channel.
	// Exchange names starting with "amq." are reserved for pre-declared and standardized exchanges.
	// The client MAY declare an exchange starting with "amq." if the passive option is set, or the exchange
	// already exists.
	// Names can consist of a non-empty sequence of letters, digits, hyphen, underscore, period, or colon.
	// Each exchange belongs to one of a set of exchange kinds/types implemented by the server.
	// The exchange types define the functionality of the exchange - i.e. how messages are routed through it.
	// Once an exchange is declared, its type cannot be changed.
	// The common types are "direct", "fanout", "topic" and "headers".
	if err := ch.ExchangeDeclare(cfg.GeofenceEventExchange, "direct", true, false, false, false, nil); err != nil {
		log.Fatalf("exchange declare failed: %v", err)
	}

	// 	QueueDeclare declares a queue to hold messages and deliver to consumers.
	// Declaring creates a queue if it doesn't already exist, or ensures that an existing
	// queue matches the same parameters.

	// Every queue declared gets a default binding to the empty exchange "" which has the type "direct"
	// with the routing key matching the queue's name. With this default binding, it is possible to
	// publish messages that route directly to this queue by publishing to ""
	// with the routing key of the queue name.
	queue, err := ch.QueueDeclare(cfg.GeofenceQueue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue declare failed: %v", err)
	}

	// QueueBind binds an exchange to a queue so that publishings to the exchange will be routed to the queue
	// when the publishing routing key matches the binding routing key.
	if err := ch.QueueBind(queue.Name, cfg.GeofenceRoutingKey, cfg.GeofenceEventExchange, false, nil); err != nil {
		log.Fatalf("queue bind failed: %v", err)
	}

	// this is rabbitmq consumer
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
