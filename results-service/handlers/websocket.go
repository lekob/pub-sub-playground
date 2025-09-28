package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"results-service/hub"
	"results-service/store"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WebSocketHandler handles WebSocket connections for real-time vote updates
type WebSocketHandler struct {
	hub   hub.ClientManager
	store store.VoteStore
}

// NewWebSocketHandler creates a new WebSocketHandler instance
func NewWebSocketHandler(h hub.ClientManager, s store.VoteStore) *WebSocketHandler {
	return &WebSocketHandler{hub: h, store: s}
}

// ServeHTTP implements the http.Handler interface
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := hub.NewWebsocketClient(conn)
	h.hub.Register(client)

	defer func() {
		h.hub.Unregister(client)
	}()

	// Send initial vote counts
	counts, err := h.store.GetVoteCounts(context.Background())
	if err != nil {
		log.Printf("Failed to get initial vote counts: %s", err)
	} else {
		if initialJSON, err := json.Marshal(counts); err == nil {
			client.WriteMessage(websocket.TextMessage, initialJSON)
		}
	}

	// Keep the connection alive until an error occurs
	for {
		if _, _, err := client.ReadMessage(); err != nil {
			break
		}
	}
}
