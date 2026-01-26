package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	// connect to RabbitMQ
	url := "amqp://guest:guest@localhost:5672/"
	conn, _ := amqp.Dial(url)
	defer conn.Close()
	fmt.Println("Successfully connected to the server")

	// declare and bind 'game_log' queue
	ch, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		pubsub.Durable,
	)
	defer ch.Close()
	if err != nil {
		fmt.Println("Error binding queue:", err)
		return
	}

	if err := pubsub.Subscribe(
		conn,
		routing.ExchangePerilTopic,
		"game_logs",
		routing.GameLogSlug+".*",
		pubsub.Durable,
		handlerGameLog(),
		pubsub.UnmarshalGob,
	); err != nil {
		fmt.Println("Failed to subscribe to game_logs queue:", err)
		return
	}

	gamelogic.PrintServerHelp()
infiniteLoop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("Pausing...")
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			}); err != nil {
				return
			}
		case "resume":
			fmt.Println("Resuming...")
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			}); err != nil {
				return
			}
		case "quit":
			fmt.Println("Quitting...")
			break infiniteLoop
		default:
			fmt.Println("Unknown command")
		}
	}

	fmt.Println("Shutting down and closing connection...")
}

func handlerGameLog() func(routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Println("> ")
		if err := gamelogic.WriteLog(gl); err != nil {
			fmt.Println("Failed to write log to disk:", err)
		}
		return pubsub.Ack
	}
}
