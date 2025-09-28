package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"results-service/consumer"
	"results-service/handlers"
	"results-service/hub"
	"results-service/store"

	"poll/common/rabbitmq"
	"poll/common/redis"
)

func main() {
	// Initialize Redis client
	redisClient, err := redis.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %s", err)
	}
	defer redisClient.Close()

	// Initialize components
	amqp, err := rabbitmq.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	var voteStore store.VoteStore = store.NewRedisStore(redisClient)
	hubInstance := hub.New()
	go hubInstance.Run()

	var messageBroadcaster hub.MessageBroadcaster = hubInstance
	var clientManager hub.ClientManager = hubInstance

	// Start the RabbitMQ consumer
	voteConsumer := consumer.New(voteStore, messageBroadcaster, amqp)
	if err := voteConsumer.Start(); err != nil {
		log.Fatalf("Failed to start vote consumer: %s", err)
	}

	// Initialize handlers
	resultsHandler := handlers.NewResultsHandler(voteStore)
	wsHandler := handlers.NewWebSocketHandler(clientManager, voteStore)
	http.Handle("/results", resultsHandler)
	http.Handle("/ws", wsHandler)

	// Set up HTTP server
	server := &http.Server{}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	server.Addr = ":" + port

	// Start service in a goroutine
	go func() {
		log.Printf("Results service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}

	log.Println("Server exiting")
}
