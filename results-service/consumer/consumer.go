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
	store store.VoteStore
	hub   *hub.Hub
	conn  *amqp.Connection
}

func New(s store.VoteStore, h *hub.Hub, c *amqp.Connection) *Consumer {
	return &Consumer{store: s, hub: h, conn: c}
}

func (c *Consumer) Start() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

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
		return err
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
		return err
	}

	go c.processMessages(msgs)

	return nil
}

func (c *Consumer) processMessages(msgs <-chan amqp.Delivery) {
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

	// If the loop exits, it means the channel was closed. This is a fatal
	// state for the service, so we should log it and exit.
	log.Fatalf("RabbitMQ consumer channel closed. Shutting down.")
}
