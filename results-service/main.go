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

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Allow all connections.
	}
	redisClient   *redis.Client
	rabbitMQQueue = "votes"
)

// Hub manages websocket clients.
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func connectToRedis() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	return rdb
}

func connectToRabbitMQ() (*amqp.Connection, error) {
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}
	var connection *amqp.Connection
	var err error
	for range 5 {
		connection, err = amqp.Dial(amqpURL)
		if err == nil {
			return connection, nil
		}
		log.Printf("Failed to connect to RabbitMQ. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	return nil, err
}

func startRabbitMQConsumer(hub *Hub) {
	conn, err := connectToRabbitMQ()
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	if queueName := os.Getenv("RABBITMQ_QUEUE"); queueName != "" {
		rabbitMQQueue = queueName
	}

	q, err := ch.QueueDeclare(
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

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	log.Println("RabbitMQ consumer started. Waiting for votes...")
	for d := range msgs {
		log.Printf("Received a vote for: %s", d.Body)
		option := string(d.Body)

		// Increment vote count in Redis.
		err := redisClient.Incr(context.Background(), option).Err()
		if err != nil {
			log.Printf("Failed to increment vote count in Redis: %s", err)
			continue
		}

		// Get updated counts and broadcast them.
		counts, err := getVoteCounts()
		if err != nil {
			log.Printf("Failed to get vote counts for broadcast: %s", err)
			continue
		}
		update, err := json.Marshal(counts)
		if err == nil {
			hub.broadcast <- update
		}
	}
}

func getVoteCounts() (map[string]int64, error) {
	ctx := context.Background()
	keys, err := redisClient.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, key := range keys {
		val, err := redisClient.Get(ctx, key).Int64()
		if err != nil {
			log.Printf("Failed to get value for key %s: %s", key, err)
			continue
		}
		counts[key] = val
	}
	return counts, nil
}

func handleResults(w http.ResponseWriter, r *http.Request) {
	counts, err := getVoteCounts()
	if err != nil {
		http.Error(w, "Failed to get results", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	hub.register <- conn

	defer func() {
		hub.unregister <- conn
	}()

	initialData, err := getVoteCounts()
	if err != nil {
		log.Printf("Failed to get initial vote counts: %s", err)
	}
	initialJSON, _ := json.Marshal(initialData)
	conn.WriteMessage(websocket.TextMessage, initialJSON)

	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}

func main() {
	// Initialize Redis client.
	redisClient = connectToRedis()
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Could not connect to Redis: %s", err)
	}
	log.Println("Successfully connected to Redis")

	// Start WebSocket hub.
	hub := newHub()
	go hub.run()

	// Start RabbitMQ consumer.
	go startRabbitMQConsumer(hub)

	// Set up HTTP server.
	server := &http.Server{Addr: ":8081"}
	http.HandleFunc("/results", handleResults)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	server.Addr = ":" + port

	// Start service in a goroutine.
	go func() {
		log.Printf("Results service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %s", err)
		}
	}()

	// Wait for a signal to gracefully shut down.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}

	log.Println("Server exiting")
}
