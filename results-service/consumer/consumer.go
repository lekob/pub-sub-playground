package consumer

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"results-service/hub"
	"results-service/store"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	store *store.Store
	hub   *hub.Hub
	conn  *amqp.Connection
}

func New(s *store.Store, h *hub.Hub, c *amqp.Connection) *Consumer {
	return &Consumer{store: s, hub: h, conn: c}
}

func (c *Consumer) Start() {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	queueName := "votes"
	if qn := os.Getenv("RABBITMQ_QUEUE"); qn != "" {
		queueName = qn
	}

	q, err := ch.QueueDeclare(
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

	ctx := context.Background()
	log.Println("RabbitMQ consumer started. Waiting for votes...")
	for d := range msgs {
		log.Printf("Received a vote for: %s", d.Body)
		option := string(d.Body)

		if err := c.store.IncrementVote(ctx, option); err != nil {
			log.Printf("Failed to increment vote count: %s", err)
			continue
		}

		counts, err := c.store.GetVoteCounts(ctx)
		if err != nil {
			log.Printf("Failed to get vote counts: %s", err)
			continue
		}

		if update, err := json.Marshal(counts); err == nil {
			c.hub.Broadcast <- update
		}
	}
}
