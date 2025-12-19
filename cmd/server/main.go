package main

import (
	"fmt"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	url := "amqp://guest:guest@localhost:5672/"
	conn, _ := amqp.Dial(url)
	defer conn.Close()
	fmt.Println("Successfully connected to the server")

	ch, _ := conn.Channel()
	defer ch.Close()

	// wait for ctrl+c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("Shutting down and closing connection...")
}
