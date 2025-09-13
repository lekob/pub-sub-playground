module polling-service

go 1.24.3

require (
	github.com/rabbitmq/amqp091-go v1.10.0
	poll/common v0.0.0-00010101000000-000000000000
)

replace poll/common => ../common
