package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"interface-api/pkg/rabbitmq"

	"github.com/skip2/go-qrcode"
)

func main() {
	rabbitURL := flag.String("url", "amqp://guest:guest@localhost:5672/", "RabbitMQ connection URL")
	queueName := flag.String("queue", "default", "Queue name to consume from")
	flag.Parse()

	consumer, err := rabbitmq.NewConsumer(*rabbitURL)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageHandler := func(body []byte) error {
		data := string(body)
		log.Printf("Received message: %s", data)

		qr, err := qrcode.New(data, qrcode.Medium)
		if err != nil {
			log.Printf("Error generating QR code: %v", err)
			return err
		}

		fmt.Println("\n" + qr.ToSmallString(false))
		fmt.Println("Message:", data)
		fmt.Println("----------------------------------------")
		return nil
	}

	err = consumer.ConsumeQueue(ctx, *queueName, messageHandler)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Printf("Connected to RabbitMQ. Consuming from queue: %s", *queueName)
	log.Println("Waiting for messages... Press Ctrl+C to exit")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("\nShutting down consumer...")
	cancel()
}
