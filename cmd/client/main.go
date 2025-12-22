package main

import (
	"fmt"
	"os"
	"os/signal"

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

	exchange := routing.ExchangePerilDirect
	queueName := routing.PauseKey + "." + username
	routingKey := routing.PauseKey
	queueType := pubsub.Transient
	ch, _, err := pubsub.DeclareAndBind(conn, exchange, queueName, routingKey, queueType)
	if err != nil {
		return
	}
	defer ch.Close()

	// wait for ctrl+c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("Shutting down and closing connection...")
}
