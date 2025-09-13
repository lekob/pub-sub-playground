package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Vote struct {
	Option string `json:"option"`
}

var (
	rabbitMQChannel *amqp.Channel
	channelMutex    sync.Mutex
	rabbitMQQueue   = "votes"
)

// connectToRabbitMQ establishes a connection and a channel to RabbitMQ.
func connectToRabbitMQ() (*amqp.Connection, error) {
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
			return connection, nil
		}
		log.Printf("Failed to connect to RabbitMQ. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	return nil, err
}

func setupRabbitMQ(connection *amqp.Connection) {
	// Get a channel.
	ch, err := connection.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %s", err)
	}

	// Get queue name from environment variable.
	if queueName := os.Getenv("RABBITMQ_QUEUE"); queueName != "" {
		rabbitMQQueue = queueName
	}

	// Create a queue.
	_, err = ch.QueueDeclare(
		rabbitMQQueue,
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
		return
	}

	// Decode request body.
	var vote Vote
	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate vote body.
	if vote.Option == "" {
		http.Error(w, "Vote option cannot be empty", http.StatusBadRequest)
		return
	}

	// Lock the channel for safe concurrent use.
	channelMutex.Lock()
	defer channelMutex.Unlock()

	// Publish the vote to the RabbitMQ queue.
	err = rabbitMQChannel.Publish(
		"",
		rabbitMQQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(vote.Option),
		},
	)
	if err != nil {
		log.Printf("Failed to publish a message: %s", err)
		http.Error(w, "Failed to process vote", http.StatusInternalServerError)
		return
	}
	log.Printf("Published vote for option: %s", vote.Option)

	// Happy path response to request.
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Vote cast successfully!"))
}

func main() {
	// Connect to RabbitMQ.
	conn, err := connectToRabbitMQ()
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ after multiple retries: %s", err)
	}
	defer conn.Close()
	log.Printf("Successfully connected to RabbitMQ")

	// Set up RabbitMQ channel and queue.
	setupRabbitMQ(conn)
	defer rabbitMQChannel.Close()

	// Set up HTTP server.
	server := &http.Server{Addr: ":8080"}
	http.HandleFunc("/vote", handleVote)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Addr = ":" + port

	// Start service in a goroutine.
	go func() {
		log.Printf("Polling service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()

	// Wait for a signal to gracefully shut down.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a context with a timeout for the shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server.
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}

	log.Println("Server exiting")
}
