package queue

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const QueueName = "ocr_processing"

type Queue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(amqpURL string) (*Queue, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao abrir channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar fila: %w", err)
	}

	return &Queue{conn: conn, channel: channel}, nil
}

func (q *Queue) Publish(ctx context.Context, body []byte) error {
	return q.channel.PublishWithContext(ctx,
		"",
		QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (q *Queue) Consume() (<-chan amqp.Delivery, error) {
	if err := q.channel.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("erro ao configurar QoS: %w", err)
	}

	return q.channel.Consume(
		QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (q *Queue) Close() {
	q.channel.Close()
	q.conn.Close()
}
