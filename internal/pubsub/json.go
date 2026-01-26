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

	if err := ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	); err != nil {
		return err
	}

	return nil
}

func UnmarshalJSON[T any](data []byte) (T, error) {
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	return value, nil
}
