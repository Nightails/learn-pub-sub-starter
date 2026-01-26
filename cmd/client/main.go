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

	moveCh, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		"army_moves."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
	)
	if err != nil {
		fmt.Println("Failed to declare and bind army_moves queue:", err)
		return
	}
	defer moveCh.Close()

	gs := gamelogic.NewGameState(username)

	// subscribe to 'pause' queue
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

	// subscribe to 'army_moves' queue
	if err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"army_moves."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerArmyMove(gs),
	); err != nil {
		fmt.Println("Failed to subscribe to army_moves queue:", err)
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
			move, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Moved successfully!")
			}
			if err := pubsub.PublishJSON(
				moveCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				move,
			); err != nil {
				fmt.Println("Failed to publish army move:", err)
			} else {
				fmt.Print("Move was published successfully.")
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerArmyMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(am gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(am)
		switch moveOutcome {
		case gamelogic.MoveOutComeSafe, gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}
