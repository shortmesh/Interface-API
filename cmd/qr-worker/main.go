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
		logger.Log.Fatalf("RabbitMQ consumer creation failed: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageHandler := func(d amqp.Delivery) error {
		data := string(d.Body)
		logger.Log.Infof("Received QR code data: %s", data)

		qr, err := qrcode.New(data, qrcode.Medium)
		if err != nil {
			logger.Log.Errorf("QR code generation failed: %v", err)
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
		logger.Log.Fatalf("Queue consumption failed: %v", err)
	}

	logger.Log.Infof("Connected to RabbitMQ. Consuming from queue: %s", *queueName)
	logger.Log.Info("Waiting for messages... Press Ctrl+C to exit")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Log.Info("Shutting down consumer")
	cancel()
}
