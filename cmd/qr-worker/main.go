package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"interface-api/pkg/logger"
	"interface-api/pkg/rabbitmq"

	"github.com/skip2/go-qrcode"
	"github.com/streadway/amqp"
)

func main() {
	rabbitURL := flag.String("url", "amqp://guest:guest@localhost:5672/", "RabbitMQ connection URL")
	queueName := flag.String("queue", "default", "Queue name to consume from")
	flag.Parse()

	consumer, err := rabbitmq.NewConsumer(*rabbitURL)
	if err != nil {
		logger.Error(fmt.Sprintf("RabbitMQ consumer creation failed: %v", err))
		os.Exit(1)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageHandler := func(d amqp.Delivery) error {
		data := string(d.Body)
		logger.Info(fmt.Sprintf("Received QR code data: %s", data))

		qr, err := qrcode.New(data, qrcode.Medium)
		if err != nil {
			logger.Error(fmt.Sprintf("QR code generation failed: %v", err))
			return err
		}

		fmt.Println("\n" + qr.ToSmallString(false))
		fmt.Println("Message:", data)
		fmt.Println("----------------------------------------")
		return nil
	}

	err = consumer.Consume(
		ctx, *queueName, messageHandler, cancel, rabbitmq.DefaultConsumeOptions(),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Queue consumption failed: %v", err))
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("Connected to RabbitMQ. Consuming from queue: %s", *queueName))
	logger.Info("Waiting for messages... Press Ctrl+C to exit")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down consumer")
	cancel()
}
