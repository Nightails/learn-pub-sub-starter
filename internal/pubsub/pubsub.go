package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}); err != nil {
		return err
	}

	return nil
}

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "Transient"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	q, err := ch.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType == Transient,
		queueType == Transient,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	if err := ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, q, nil
}
