package rabbitmq

import (
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Connect establishes a connection to RabbitMQ and returns the connection object.
func Connect() (*amqp.Connection, error) {
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	var connection *amqp.Connection
	var err error
	for range 5 {
		connection, err = amqp.Dial(amqpURL)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ")
			return connection, nil
		}
		log.Printf("Failed to connect to RabbitMQ. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	return nil, err
}
