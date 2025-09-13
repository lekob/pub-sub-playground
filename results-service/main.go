package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// VoteCounts holds the counts for each option.
// We use a sync.RWMutex to handle concurrent reads/writes safely.
var VoteCounts = struct {
	sync.RWMutex
	Counts map[string]int
}{Counts: make(map[string]int)}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all connections.
}

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

var hub = newHub()

func handleResults(w http.ResponseWriter, r *http.Request) {
	VoteCounts.RLock()
	defer VoteCounts.RUnlock()
	json.NewEncoder(w).Encode(VoteCounts.Counts)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	hub.register <- conn

	// Make sure that the client is unregistered when the connection is closed.
	defer func() {
		hub.unregister <- conn
	}()

	VoteCounts.RLock()
	initialData, _ := json.Marshal(VoteCounts.Counts)
	conn.WriteMessage(websocket.TextMessage, initialData)

	// Keep the connection alive. The Hub will broadcast.
	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}

func startRabbitMQConsumer() {
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
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

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	log.Println("RabbitMQ consumer started. Waiting for votes...")
	for d := range msgs {
		log.Printf("Received a vote for: %s", d.Body)
		option := string(d.Body)

		// Safely update the vote counts
		VoteCounts.Lock()
		VoteCounts.Counts[option]++
		VoteCounts.Unlock()

		// Get the current counts and broadcast them
		VoteCounts.RLock()
		updateCounts, err := json.Marshal(VoteCounts.Counts)
		VoteCounts.RUnlock()
		if err == nil {
			hub.broadcast <- updateCounts
		}
	}
}

func main() {
	// Start the WebSocket hub.
	go hub.run()

	// Start the RabbitMQ consumer.
	go startRabbitMQConsumer()

	http.HandleFunc("/results", handleResults)
	http.HandleFunc("/ws", handleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Results service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
