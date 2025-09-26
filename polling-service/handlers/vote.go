package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Vote struct {
	Option string `json:"option"`
}

type VoteHandler struct {
	rabbitMQChannel *amqp.Channel
	rabbitMQQueue   string
	channelMutex    sync.Mutex
}

func NewVoteHandler(ch *amqp.Channel, queue string) *VoteHandler {
	return &VoteHandler{
		rabbitMQChannel: ch,
		rabbitMQQueue:   queue,
	}
}

func (h *VoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var vote Vote
	if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if vote.Option == "" {
		http.Error(w, "Vote option cannot be empty", http.StatusBadRequest)
		return
	}

	h.channelMutex.Lock()
	defer h.channelMutex.Unlock()

	err := h.rabbitMQChannel.Publish(
		"",
		h.rabbitMQQueue,
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
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Vote cast successfully!"))
}
