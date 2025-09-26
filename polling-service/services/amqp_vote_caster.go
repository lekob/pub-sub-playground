package services

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpVoteCaster struct {
	amqpChannel  *amqp.Channel
	amqpQueue    string
	channelMutex sync.Mutex
}

func NewAmqpVoteCaster(ch *amqp.Channel, queue string) *AmqpVoteCaster {
	return &AmqpVoteCaster{
		amqpChannel: ch,
		amqpQueue:   queue,
	}
}

func (s *AmqpVoteCaster) CastVote(option string) error {
	s.channelMutex.Lock()
	defer s.channelMutex.Unlock()

	return s.amqpChannel.Publish(
		"",
		s.amqpQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(option),
		},
	)
}
