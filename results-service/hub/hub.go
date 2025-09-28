package hub

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type hub struct {
	clients    map[Client]bool
	broadcast  chan []byte
	register   chan Client
	unregister chan Client
	mu         sync.Mutex
}

func New() Hub {
	return &hub{
		broadcast:  make(chan []byte),
		register:   make(chan Client),
		unregister: make(chan Client),
		clients:    make(map[Client]bool),
	}
}

func (h *hub) Register(client Client) {
	h.register <- client
}

func (h *hub) Unregister(client Client) {
	h.unregister <- client
}

func (h *hub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (h *hub) Run() {
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
			// Make a copy of the clients map under the lock to avoid holding the
			// lock during the potentially slow WriteMessage operation.
			clientsToBroadcast := make([]Client, 0, len(h.clients))
			h.mu.Lock()
			for client := range h.clients {
				clientsToBroadcast = append(clientsToBroadcast, client)
			}
			h.mu.Unlock()

			for _, client := range clientsToBroadcast {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("error writing to client: %v", err)
					h.Unregister(client) // Use the unregister channel to handle cleanup
				}
			}
		}
	}
}
