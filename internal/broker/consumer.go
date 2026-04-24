package broker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQ) GetTrackTaskStream() (<-chan amqp.Delivery, error) {
	const op = "broker.GetTrackTaskStream"

	err := r.Channel.Qos(1, 0, false)
	if err != nil {
		return nil, fmt.Errorf("%s: can't set QoS: %w", op, err)
	}

	msgs, err := r.Channel.Consume(
		AudioTasksQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: can't consume messages: %w", op, err)
	}

	return msgs, nil
}
