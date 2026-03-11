package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"interface-api/pkg/config"
	"interface-api/pkg/logger"
	"interface-api/pkg/matrixclient"
	"interface-api/pkg/rabbitmq"
	"interface-api/pkg/throttler"

	"github.com/streadway/amqp"
)

type QueuedMessage struct {
	DeviceID     string `json:"device_id"`
	Contact      string `json:"contact"`
	PlatformName string `json:"platform_name"`
	Text         string `json:"text"`
	Username     string `json:"username"`
}

type Worker struct {
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	workerCount     int
	rabbitURL       string
	exchangeName    string
	queueName       string
	delayQueueName  string
	sharedThrottler *throttler.Throttler
}

func New() *Worker {
	workerCount := 1
	if count := os.Getenv("WORKER_COUNT"); count != "" {
		if n, err := strconv.Atoi(count); err == nil && n > 0 {
			workerCount = n
		}
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	if err := config.ValidateExternalURL(rabbitURL, "RABBITMQ_URL"); err != nil {
		logger.Warn(fmt.Sprintf("RabbitMQ URL validation warning: %v", err))
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

	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		ctx:             ctx,
		cancel:          cancel,
		workerCount:     workerCount,
		rabbitURL:       rabbitURL,
		exchangeName:    exchangeName,
		queueName:       queueName,
		delayQueueName:  delayQueueName,
		sharedThrottler: throttler.New(),
	}
}

func IsEnabled() bool {
	return os.Getenv("WORKER_ENABLED") != "false"
}

func (w *Worker) Start() {
	logger.Info(fmt.Sprintf("Starting %d message worker(s)", w.workerCount))

	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go func(workerID int) {
			defer w.wg.Done()
			w.runWorker(workerID)
		}(i + 1)
	}
}

func (w *Worker) Stop() {
	logger.Info("Shutting down workers")
	w.cancel()
	w.wg.Wait()
	logger.Info("All workers stopped")
}

func (w *Worker) runWorker(workerID int) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("Worker %d panic: %v\n%s", workerID, r, debug.Stack()))
		}
	}()

	logger.Info(fmt.Sprintf("Worker %d starting", workerID))

	matrixClient, err := matrixclient.New()
	if err != nil {
		logger.Error(fmt.Sprintf("Worker %d: Matrix client initialization failed: %v", workerID, err))
		return
	}

	retryDelay := 1 * time.Second
	maxRetryDelay := 60 * time.Second
	connectionSuccessThreshold := 30 * time.Second

	for {
		select {
		case <-w.ctx.Done():
			logger.Info(fmt.Sprintf("Worker %d: Shutting down", workerID))
			return
		default:
		}

		logger.Info(fmt.Sprintf("Worker %d: Connecting to RabbitMQ", workerID))
		connectionStart := time.Now()
		err := w.runWorkerLoop(workerID, matrixClient)
		connectionDuration := time.Since(connectionStart)

		if err == nil {
			logger.Info(fmt.Sprintf("Worker %d: Connection closed cleanly", workerID))
			return
		}

		if connectionDuration >= connectionSuccessThreshold {
			retryDelay = 1 * time.Second
		}

		select {
		case <-w.ctx.Done():
			logger.Info(fmt.Sprintf("Worker %d: Shutting down", workerID))
			return
		default:
			logger.Warn(fmt.Sprintf("Worker %d: Connection lost after %v, reconnecting in %v", workerID, connectionDuration.Round(time.Second), retryDelay))
			time.Sleep(retryDelay)
			retryDelay *= 2
			if retryDelay > maxRetryDelay {
				retryDelay = maxRetryDelay
			}
		}
	}
}

func (w *Worker) runWorkerLoop(workerID int, matrixClient *matrixclient.Client) error {
	ctx, cancel := context.WithCancel(w.ctx)
	defer cancel()

	consumer, err := rabbitmq.NewConsumer(w.rabbitURL)
	if err != nil {
		return err
	}
	defer consumer.Close()

	producer, err := rabbitmq.NewProducer(w.rabbitURL)
	if err != nil {
		return err
	}
	defer producer.Close()

	delayQueueArgs := amqp.Table{
		"x-dead-letter-exchange":    w.exchangeName,
		"x-dead-letter-routing-key": "message.*.*",
	}
	delayQueueConfig := rabbitmq.DefaultQueueConfig(w.delayQueueName)
	delayQueueConfig.Args = delayQueueArgs

	if err := consumer.DeclareQueue(delayQueueConfig); err != nil {
		return err
	}

	deliveryHandler := func(delivery amqp.Delivery) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("Worker %d: Message handler panic: %v\n%s", workerID, r, debug.Stack()))
				delivery.Nack(false, false)
			}
		}()

		var msg QueuedMessage
		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			logger.Error(fmt.Sprintf("Worker %d: Message unmarshal failed: %v", workerID, err))
			delivery.Nack(false, false)
			return err
		}

		if !w.sharedThrottler.Allow(msg.PlatformName, msg.Username) {
			waitTime := w.sharedThrottler.WaitTime(msg.PlatformName, msg.Username)
			logger.Info(fmt.Sprintf("Worker %d: Rate limit applied, delaying %v", workerID, waitTime))

			publishOpts := rabbitmq.DefaultPublishOptions()
			publishOpts.Expiration = fmt.Sprintf("%d", waitTime.Milliseconds())

			if err := producer.Publish("", w.delayQueueName, msg, publishOpts); err != nil {
				logger.Error(fmt.Sprintf("Worker %d: Delay queue publish failed: %v\n%s", workerID, err, debug.Stack()))
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
			logger.Error(fmt.Sprintf("Worker %d: Message delivery failed: %v", workerID, err))
			delivery.Nack(false, false)
			return err
		}

		logger.Info(fmt.Sprintf("Worker %d: Message delivered successfully", workerID))
		delivery.Ack(false)
		return nil
	}

	opts := rabbitmq.ManualAckOptions()
	opts.CreateQueue = true
	opts.CreateExchange = true
	opts.ExchangeType = "topic"
	opts.BindExchange = w.exchangeName
	opts.BindingKey = "message.*.*"

	err = consumer.Consume(ctx, w.queueName, deliveryHandler, cancel, opts)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Worker %d: Connected and listening on exchange '%s' with pattern 'message.*.*'", workerID, w.exchangeName))

	<-ctx.Done()

	select {
	case <-w.ctx.Done():
		return nil
	default:
		return fmt.Errorf("connection closed")
	}
}
