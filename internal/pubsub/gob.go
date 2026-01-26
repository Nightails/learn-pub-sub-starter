package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, value T) error {
	buffer := new(bytes.Buffer)
	if err := gob.NewEncoder(buffer).Encode(value); err != nil {
		return err
	}

	if err := ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        buffer.Bytes(),
		},
	); err != nil {
		return err
	}

	return nil
}

func UnmarshalGob[T any](data []byte) (T, error) {
	var value T
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&value); err != nil {
		return value, err
	}
	return value, nil
}
