package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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
	DeviceID      string `json:"device_id"`
	Contact       string `json:"contact"`
	PlatformName  string `json:"platform_name"`
	Text          string `json:"text"`
	Username      string `json:"username"`
	FileContent   string `json:"file_content,omitempty"`
	FileExtension string `json:"file_extension,omitempty"`
	GroupURL      string `json:"group_url,omitempty"`
	ReplyID       string `json:"reply_id,omitempty"`
}

const (
	RetryCountHeader       = "x-retry-count"
	MessageCreatedAtHeader = "x-message-created-at"
)

type Worker struct {
	ctx                      context.Context
	cancel                   context.CancelFunc
	wg                       sync.WaitGroup
	workerCount              int
	rabbitURL                string
	exchangeName             string
	queueName                string
	delayQueueName           string
	retryQueueName           string
	initialRetryDelay        time.Duration
	exponentialBackoffFactor float64
	exponentialBackoffStart  int64
	maxRetryDelay            time.Duration
	messageMaxTTL            time.Duration
	sharedThrottler          *throttler.Throttler
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

	retryQueueName := os.Getenv("MESSAGE_RETRY_QUEUE_NAME")
	if retryQueueName == "" {
		retryQueueName = "shortmesh-messages-retry-queue"
	}

	initialRetryDelay := 5 * time.Second
	if delayStr := os.Getenv("MESSAGE_INITIAL_RETRY_DELAY_SECONDS"); delayStr != "" {
		if seconds, err := strconv.Atoi(delayStr); err == nil && seconds > 0 {
			initialRetryDelay = time.Duration(seconds) * time.Second
		}
	}

	exponentialBackoffFactor := 2.0
	if factorStr := os.Getenv("MESSAGE_EXPONENTIAL_BACKOFF_FACTOR"); factorStr != "" {
		if factor, err := strconv.ParseFloat(factorStr, 64); err == nil && factor > 1.0 {
			exponentialBackoffFactor = factor
		}
	}

	exponentialBackoffStart := int64(5)
	if startStr := os.Getenv("MESSAGE_EXPONENTIAL_BACKOFF_START"); startStr != "" {
		if start, err := strconv.ParseInt(startStr, 10, 64); err == nil && start > 0 {
			exponentialBackoffStart = start
		}
	}

	maxRetryDelay := 5 * time.Minute
	if delayStr := os.Getenv("MESSAGE_MAX_RETRY_DELAY_MINUTES"); delayStr != "" {
		if minutes, err := strconv.Atoi(delayStr); err == nil && minutes > 0 {
			maxRetryDelay = time.Duration(minutes) * time.Minute
		}
	}

	messageMaxTTL := 24 * time.Hour
	if ttlStr := os.Getenv("MESSAGE_MAX_TTL_HOURS"); ttlStr != "" {
		if hours, err := strconv.Atoi(ttlStr); err == nil && hours > 0 {
			messageMaxTTL = time.Duration(hours) * time.Hour
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		ctx:                      ctx,
		cancel:                   cancel,
		workerCount:              workerCount,
		rabbitURL:                rabbitURL,
		exchangeName:             exchangeName,
		queueName:                queueName,
		delayQueueName:           delayQueueName,
		retryQueueName:           retryQueueName,
		initialRetryDelay:        initialRetryDelay,
		exponentialBackoffFactor: exponentialBackoffFactor,
		exponentialBackoffStart:  exponentialBackoffStart,
		maxRetryDelay:            maxRetryDelay,
		messageMaxTTL:            messageMaxTTL,
		sharedThrottler:          throttler.New(),
	}
}

func (w *Worker) calculateRetryDelay(retryCount int64) time.Duration {
	if retryCount < w.exponentialBackoffStart {
		return w.initialRetryDelay
	}

	exponent := retryCount - w.exponentialBackoffStart
	delay := float64(w.initialRetryDelay) * math.Pow(w.exponentialBackoffFactor, float64(exponent))

	result := time.Duration(delay)
	if result > w.maxRetryDelay {
		return w.maxRetryDelay
	}

	return result
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

	retryQueueArgs := amqp.Table{
		"x-dead-letter-exchange":    w.exchangeName,
		"x-dead-letter-routing-key": "message.*.*",
	}
	retryQueueConfig := rabbitmq.DefaultQueueConfig(w.retryQueueName)
	retryQueueConfig.Args = retryQueueArgs

	if err := consumer.DeclareQueue(retryQueueConfig); err != nil {
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

		retryCount := int64(0)
		messageCreatedAt := time.Now()
		if delivery.Headers != nil {
			if count, ok := delivery.Headers[RetryCountHeader].(int64); ok {
				retryCount = count
			}
			if createdAtStr, ok := delivery.Headers[MessageCreatedAtHeader].(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
					messageCreatedAt = parsedTime
				}
			}
		}

		messageAge := time.Since(messageCreatedAt)
		if messageAge > w.messageMaxTTL {
			logger.Warn(fmt.Sprintf("Worker %d: Message exceeded max TTL (%v), discarding", workerID, w.messageMaxTTL))
			delivery.Ack(false)
			return nil
		}

		remainingTTL := w.messageMaxTTL - messageAge

		if !w.sharedThrottler.Allow(msg.PlatformName, msg.Username) {
			waitTime := w.sharedThrottler.WaitTime(msg.PlatformName, msg.Username)
			logger.Info(fmt.Sprintf("Worker %d: Rate limit applied, delaying %v", workerID, waitTime))

			publishOpts := rabbitmq.DefaultPublishOptions()
			publishOpts.Expiration = fmt.Sprintf("%d", waitTime.Milliseconds())
			publishOpts.Headers = amqp.Table{
				RetryCountHeader:       retryCount,
				MessageCreatedAtHeader: messageCreatedAt.Format(time.RFC3339),
			}

			if err := producer.Publish("", w.delayQueueName, msg, publishOpts); err != nil {
				logger.Error(fmt.Sprintf("Worker %d: Delay queue publish failed: %v\n%s", workerID, err, debug.Stack()))
				delivery.Nack(false, true)
				return err
			}

			delivery.Ack(false)
			return nil
		}

		req := &matrixclient.SendMessageRequest{
			Contact:       msg.Contact,
			PlatformName:  msg.PlatformName,
			Text:          msg.Text,
			Username:      msg.Username,
			FileContent:   msg.FileContent,
			FileExtension: msg.FileExtension,
			GroupURL:      msg.GroupURL,
			ReplyID:       msg.ReplyID,
		}

		_, err := matrixClient.SendMessage(msg.DeviceID, req)
		if err != nil {
			logger.Error(fmt.Sprintf("Worker %d: Message delivery failed (attempt %d): %v", workerID, retryCount+1, err))

			retryCount++
			retryDelay := w.calculateRetryDelay(retryCount)

			if retryDelay > remainingTTL {
				logger.Warn(fmt.Sprintf("Worker %d: Retry delay (%v) exceeds remaining TTL (%v), discarding message", workerID, retryDelay, remainingTTL))
				delivery.Ack(false)
				return nil
			}

			logger.Info(fmt.Sprintf("Worker %d: Scheduling retry #%d with delay %v", workerID, retryCount, retryDelay))

			publishOpts := rabbitmq.DefaultPublishOptions()
			publishOpts.Expiration = fmt.Sprintf("%d", retryDelay.Milliseconds())
			publishOpts.Headers = amqp.Table{
				RetryCountHeader:       retryCount,
				MessageCreatedAtHeader: messageCreatedAt.Format(time.RFC3339),
			}

			if err := producer.Publish("", w.retryQueueName, msg, publishOpts); err != nil {
				logger.Error(fmt.Sprintf("Worker %d: Retry queue publish failed: %v", workerID, err))
				delivery.Nack(false, true)
				return err
			}

			delivery.Ack(false)
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
