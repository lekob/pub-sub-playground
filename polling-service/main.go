package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Vote struct {
	Option string `json:"option"`
}

var rabbitMQChannel *amqp.Channel

// connectToRabbitMQ establishes a connection and a channel to RabbitMQ.
func connectToRabbitMQ() {
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

	rabbitMQChannel = ch
}

func handleVote(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests.
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	// Decode request body.
	var vote Vote
	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	// Validate vote body.
	if vote.Option == "" {
		http.Error(w, "Vote option cannot be empty", http.StatusBadRequest)
	}

	// Publish the vote to the RabbitMQ queue.
	err = rabbitMQChannel.Publish(
		"",
		"votes",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(vote.Option),
		},
	)
	if err != nil {
		log.Printf("Faild to publish a message: %s", err)
		http.Error(w, "Failed to process vote", http.StatusInternalServerError)
	}
	log.Printf("Published vote for option: %s", vote.Option)

	// Happy path response to request.
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Vote cast successfully!"))
}

func main() {
	// Connect to RabbitMQ.
	connectToRabbitMQ()

	// Set up HTTP server.
	http.HandleFunc("/vote", handleVote)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start service.
	log.Printf("Polling service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server %s", err)
	}
}
