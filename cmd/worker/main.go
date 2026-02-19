package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"interface-api/pkg/logger"
	"interface-api/pkg/matrixclient"
	"interface-api/pkg/rabbitmq"
	"interface-api/pkg/throttler"

	_ "github.com/joho/godotenv/autoload"
	"github.com/streadway/amqp"
)

type QueuedMessage struct {
	DeviceID     string `json:"device_id"`
	Contact      string `json:"contact"`
	PlatformName string `json:"platform_name"`
	Text         string `json:"text"`
	Username     string `json:"username"`
}

func main() {
	workerCount := flag.Int("n", 1, "Number of concurrent workers")
	flag.Parse()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	exchangeName := os.Getenv("MESSAGE_EXCHANGE_NAME")
	if exchangeName == "" {
		exchangeName = "shortmesh.messages"
	}

	queueName := os.Getenv("MESSAGE_QUEUE_NAME")
	if queueName == "" {
		queueName = "shortmesh-messages-queue"
	}

	delayQueueName := os.Getenv("MESSAGE_DELAY_QUEUE_NAME")
	if delayQueueName == "" {
		delayQueueName = "shortmesh-messages-delay-queue"
	}

	logger.Log.Infof("Starting message worker service with %d workers", *workerCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sharedThrottler := throttler.New()

	var wg sync.WaitGroup

	for i := 0; i < *workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runWorker(ctx, workerID, rabbitURL, exchangeName, queueName, delayQueueName, sharedThrottler)
		}(i + 1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Log.Info("Shutting down workers")
	cancel()
	wg.Wait()
	logger.Log.Info("All workers stopped")
}

func runWorker(parentCtx context.Context, workerID int, rabbitURL string, exchangeName string, queueName string, delayQueueName string, thr *throttler.Throttler) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("Worker %d panic: %v\n%s", workerID, r, debug.Stack())
		}
	}()

	logger.Log.Infof("Worker %d starting", workerID)

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	matrixClient, err := matrixclient.New()
	if err != nil {
		logger.Log.Errorf("Worker %d: Matrix client initialization failed: %v", workerID, err)
		return
	}

	consumer, err := rabbitmq.NewConsumer(rabbitURL)
	if err != nil {
		logger.Log.Errorf("Worker %d: RabbitMQ consumer initialization failed: %v", workerID, err)
		return
	}
	defer consumer.Close()

	producer, err := rabbitmq.NewProducer(rabbitURL)
	if err != nil {
		logger.Log.Errorf("Worker %d: RabbitMQ producer initialization failed: %v", workerID, err)
		return
	}
	defer producer.Close()

	delayQueueArgs := amqp.Table{
		"x-dead-letter-exchange":    exchangeName,
		"x-dead-letter-routing-key": "message.*.*",
	}
	delayQueueConfig := rabbitmq.DefaultQueueConfig(delayQueueName)
	delayQueueConfig.Args = delayQueueArgs

	if err := consumer.DeclareQueue(delayQueueConfig); err != nil {
		logger.Log.Errorf("Worker %d: Delay queue declaration failed: %v", workerID, err)
		return
	}

	deliveryHandler := func(delivery amqp.Delivery) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Errorf("Worker %d: Message handler panic: %v\n%s", workerID, r, debug.Stack())
				delivery.Nack(false, false)
			}
		}()

		var msg QueuedMessage
		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			logger.Log.Errorf("Worker %d: Message unmarshal failed: %v", workerID, err)
			delivery.Nack(false, false)
			return err
		}

		if !thr.Allow(msg.PlatformName, msg.Username) {
			waitTime := thr.WaitTime(msg.PlatformName, msg.Username)
			logger.Log.Infof("Worker %d: Rate limit applied, delaying %v", workerID, waitTime)

			publishOpts := rabbitmq.DefaultPublishOptions()
			publishOpts.Expiration = fmt.Sprintf("%d", waitTime.Milliseconds())

			if err := producer.Publish("", delayQueueName, msg, publishOpts); err != nil {
				logger.Log.Errorf("Worker %d: Delay queue publish failed: %v\n%s", workerID, err, debug.Stack())
				delivery.Nack(false, true)
				return err
			}

			delivery.Ack(false)
			return nil
		}

		req := &matrixclient.SendMessageRequest{
			Contact:      msg.Contact,
			PlatformName: msg.PlatformName,
			Text:         msg.Text,
			Username:     msg.Username,
		}

		_, err := matrixClient.SendMessage(msg.DeviceID, req)
		if err != nil {
			logger.Log.Errorf("Worker %d: Message delivery failed: %v", workerID, err)
			delivery.Nack(false, true)
			return err
		}

		logger.Log.Infof("Worker %d: Message delivered successfully", workerID)
		delivery.Ack(false)
		return nil
	}

	opts := rabbitmq.ManualAckOptions()
	opts.CreateQueue = true
	opts.CreateExchange = true
	opts.ExchangeType = "topic"
	opts.BindExchange = exchangeName
	opts.BindingKey = "message.*.*"

	err = consumer.Consume(ctx, queueName, deliveryHandler, cancel, opts)
	if err != nil {
		logger.Log.Errorf("Worker %d: Queue consumption failed: %v", workerID, err)
		return
	}

	logger.Log.Infof("Worker %d: Listening for messages on exchange '%s' with pattern 'message.*.*'", workerID, exchangeName)

	<-ctx.Done()
	logger.Log.Infof("Worker %d: Shutting down", workerID)
}
