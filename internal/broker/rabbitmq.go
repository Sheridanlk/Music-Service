package broker

import (
	"fmt"
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func New(user, password, host string, port int) (*RabbitMQ, error) {
	const op = "broker.rabbitmq.New"

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

	return &RabbitMQ{
		Conn:    conn,
		Channel: ch,
	}, nil
}

func (r *RabbitMQ) Close() {
	r.Channel.Close()
	r.Conn.Close()
}
