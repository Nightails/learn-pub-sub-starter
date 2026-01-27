package pubsub

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nack_requeue"
	NackDiscard AckType = "nack_discard"
)

func Subscribe[T any](
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	unmarshal func([]byte) (T, error),
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}
	go func(subChan *amqp.Channel, msgsChan <-chan amqp.Delivery) {
		defer subChan.Close()
		for msg := range msgsChan {
			value, err := unmarshal(msg.Body)
			if err != nil {
				if err := msg.Nack(false, false); err != nil {
					log.Fatalf("Failed to nack message: %v", err)
				}
				continue
			}
			acktype := handler(value)
			switch acktype {
			case Ack:
				if err := msg.Ack(false); err != nil {
					log.Fatalf("Failed to ack message: %v", err)
					return
				}
			case NackRequeue:
				if err := msg.Nack(false, true); err != nil {
					log.Fatalf("Failed to nack message: %v", err)
					return
				}
			case NackDiscard:
				if err := msg.Nack(false, false); err != nil {
					log.Fatalf("Failed to nack message: %v", err)
					return
				}
			default:
				if err := msg.Nack(false, false); err != nil {
					log.Fatalf("Failed to nack message: %v", err)
					return
				}
			}
		}
	}(ch, msgs)
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
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	if err := ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, q, nil
}
