package broker

import (
	"fmt"
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
)

const AudioTasksQueue = "audio_tasks"

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func New(user, password, host string, port int) (*RabbitMQ, error) {
	const op = "broker.New"

	connString := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
	}

	conn, err := amqp.Dial(connString.String())
	if err != nil {
		return nil, fmt.Errorf("%s: can't connect to RabbitMQ: %w", op, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("%s: can't create channel: %w", op, err)
	}

	r := &RabbitMQ{
		Conn:    conn,
		Channel: ch,
	}

	if err := r.initQueue(); err != nil {
		r.Close()
		return nil, fmt.Errorf("%s: can't initialize queue: %w", op, err)
	}

	return r, nil
}

func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}

func (r *RabbitMQ) initQueue() error {
	_, err := r.Channel.QueueDeclare(
		AudioTasksQueue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	return err
}
