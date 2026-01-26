package pubsub

import (
	"context"
	"encoding/json"
	"log"

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

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nack_requeue"
	NackDiscard AckType = "nack_discard"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
	handler func(T) AckType) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	msg, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		defer ch.Close()
		for m := range msg {
			var value T
			if err := json.Unmarshal(m.Body, &value); err != nil {
				continue
			}
			acktype := handler(value)
			switch acktype {
			case Ack:
				log.Println("Acknowledged message")
				if err := m.Ack(true); err != nil {
					log.Fatalf("Failed to ack message: %v", err)
					return
				}
				log.Println("Message acknowledged")
			case NackRequeue:
				log.Println("Nacking message, requeueing")
				if err := m.Nack(false, true); err != nil {
					log.Fatalf("Failed to nack message: %v", err)
					return
				}
				log.Println("Message nacked, requeued")
			case NackDiscard:
				log.Println("Nacking message, discarding")
				if err := m.Nack(true, false); err != nil {
					log.Fatalf("Failed to nack message: %v", err)
					return
				}
				log.Println("Message nacked, discarded")
			}
		}
	}()
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
