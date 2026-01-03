package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	url := "amqp://guest:guest@localhost:5672/"
	conn, _ := amqp.Dial(url)
	defer conn.Close()
	fmt.Println("Successfully connected to the server")

	// prompt for username
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		return
	}

	// declare 'pause' queue
	exchange := routing.ExchangePerilDirect
	queueName := routing.PauseKey + "." + username
	routingKey := routing.PauseKey
	queueType := pubsub.Transient
	ch, _, err := pubsub.DeclareAndBind(conn, exchange, queueName, routingKey, queueType)
	if err != nil {
		return
	}
	defer ch.Close()

	gs := gamelogic.NewGameState(username)

	if err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		"pause."+username,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gs),
	); err != nil {
		fmt.Println("Failed to subscribe to pause queue:", err)
		return
	}

infiniteLoop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			if err := gs.CommandSpawn(words); err != nil {
				fmt.Println(err)
			}
		case "move":
			if _, err := gs.CommandMove(words); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Moved successfully!")
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			fmt.Println("Quitting...")
			break infiniteLoop
		default:
			fmt.Println("Unknown command")
		}
	}

	fmt.Println("Shutting down and closing connection...")
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
