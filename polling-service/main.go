package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"poll/common/rabbitmq"
	"polling-service/handlers"
)

func main() {
	// Connect to RabbitMQ.
	conn, err := rabbitmq.Connect()
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	// Set up RabbitMQ channel and queue.
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %s", err)
	}
	defer ch.Close()

	queueName := "votes"
	if qn := os.Getenv("RABBITMQ_QUEUE"); qn != "" {
		queueName = qn
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	// Set up HTTP server.
	server := &http.Server{}
	voteHandler := handlers.NewVoteHandler(ch, queueName)
	http.Handle("/vote", voteHandler)

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
