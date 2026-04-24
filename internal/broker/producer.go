package broker

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQ) SendTrackTask(ctx context.Context, trackId string) error {
	const op = "broker.SendTrackTask"

	err := r.Channel.PublishWithContext(
		ctx,
		"",              // exchange
		AudioTasksQueue, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(trackId),
		},
	)
	if err != nil {
		return fmt.Errorf("%s: can't publish message: %w", op, err)
	}

	return nil
}
