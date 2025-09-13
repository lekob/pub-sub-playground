package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"time"
)

// connectToRabbitMQ establishes a connection and a channel to RabbitMQ.
func connectToRabbitMQ() (*amqp.Connection, *amqp.Channel) {
	// Get RabbitMQ URL.
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	// Connect to RabbitMQ.
	var connection *amqp.Connection
	var err error
	for i := 0; i < 5; i++ {
		connection, err = amqp.Dial(amqpURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ after multiple retries: %s", err)
	}
	log.Printf("Successfuly connected to RabbitMQ")

	// Get a channel.
	ch, err := connection.Channel()
	if err != nil {
		log.Fatalf("Failed to opne chanel: %s", err)
	}

	// Create a queue.
	_, err = ch.QueueDeclare(
		"votes",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	return connection, ch
}

func main() {
	conn, ch := connectToRabbitMQ()
	defer conn.Close()
	defer ch.Close()

	// Consume messages in the "votes" queue.
	msgs, err := ch.Consume(
		"votes",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	// This channel will be blocked forever, to keep this service alive.
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("Recieved a vote for: %s", d.Body)
			d.Ack(false)
		}
	}()
	log.Printf("Worker service started. Waiting for votes...")
	// Block forever.
	<-forever
}
