package services

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type VoteService struct {
	amqpChannel  *amqp.Channel
	amqpQueue    string
	channelMutex sync.Mutex
}

func NewVoteService(ch *amqp.Channel, queue string) *VoteService {
	return &VoteService{
		amqpChannel: ch,
		amqpQueue:   queue,
	}
}

func (s *VoteService) CastVote(option string) error {
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
