package webhookworker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/config"
	"interface-api/pkg/logger"
	"interface-api/pkg/rabbitmq"

	"github.com/streadway/amqp"
)

type MediaInfo struct {
	Size     float64 `json:"Size"`
	MimeType string  `json:"MimeType"`
	Width    int     `json:"Width"`
	Height   int     `json:"Height"`
	BlurHash string  `json:"BlurHash"`
}

type Media struct {
	Content []byte    `json:"Content"`
	Info    MediaInfo `json:"Info"`
}

type IncomingMessage struct {
	IsContact bool   `json:"IsContact"`
	Type      string `json:"Type"`
	From      string `json:"From"`
	To        string `json:"To"`
	Media     Media  `json:"Media"`
}

type userConsumer struct {
	matrixUsername string
	cancel         context.CancelFunc
}

type WebhookWorker struct {
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
	rabbitURL          string
	exchangeName       string
	bindingKey         string
	db                 database.Service
	userConsumers      map[string]*userConsumer
	userConsumersMutex sync.RWMutex
	refreshInterval    time.Duration
}

func New(db database.Service) *WebhookWorker {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	if err := config.ValidateExternalURL(rabbitURL, "RABBITMQ_URL"); err != nil {
		logger.Warn(fmt.Sprintf("RabbitMQ URL validation warning: %v", err))
	}

	exchangeName := os.Getenv("WEBHOOK_INCOMING_EXCHANGE")
	if exchangeName == "" {
		exchangeName = "contacts.topic"
	}

	bindingKey := os.Getenv("WEBHOOK_INCOMING_BINDING_KEY")
	if bindingKey == "" {
		bindingKey = "contacts.topic.incoming_messages"
	}

	refreshInterval := 30 * time.Second
	if interval := os.Getenv("WEBHOOK_REFRESH_INTERVAL_SECONDS"); interval != "" {
		if seconds, err := strconv.Atoi(interval); err == nil && seconds > 0 {
			refreshInterval = time.Duration(seconds) * time.Second
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WebhookWorker{
		ctx:             ctx,
		cancel:          cancel,
		rabbitURL:       rabbitURL,
		exchangeName:    exchangeName,
		bindingKey:      bindingKey,
		db:              db,
		userConsumers:   make(map[string]*userConsumer),
		refreshInterval: refreshInterval,
	}
}

func IsEnabled() bool {
	return os.Getenv("WEBHOOK_WORKER_ENABLED") != "false"
}

func (w *WebhookWorker) Start() {
	logger.Info("Starting webhook worker manager")

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.manageConsumers()
	}()
}

func (w *WebhookWorker) Stop() {
	logger.Info("Shutting down webhook worker")
	w.cancel()
	w.wg.Wait()
	logger.Info("Webhook worker stopped")
}

func (w *WebhookWorker) manageConsumers() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("Webhook worker manager panic: %v\n%s", r, debug.Stack()))
		}
	}()

	ticker := time.NewTicker(w.refreshInterval)
	defer ticker.Stop()

	w.syncConsumers()

	for {
		select {
		case <-w.ctx.Done():
			logger.Info("Webhook worker manager: Shutting down")
			return
		case <-ticker.C:
			w.syncConsumers()
		}
	}
}

func (w *WebhookWorker) syncConsumers() {
	identities, err := w.getActiveMatrixIdentities()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to fetch matrix identities: %v", err))
		return
	}

	usernameMap := make(map[string]uint)
	for _, identity := range identities {
		if _, exists := usernameMap[identity.MatrixUsername]; !exists {
			usernameMap[identity.MatrixUsername] = identity.ID
		}
	}

	w.userConsumersMutex.Lock()

	for username, consumer := range w.userConsumers {
		if _, exists := usernameMap[username]; !exists {
			logger.Info("Stopping consumer for removed user")
			consumer.cancel()
			delete(w.userConsumers, username)
		}
	}

	for username, identityID := range usernameMap {
		if _, exists := w.userConsumers[username]; !exists {
			logger.Info("Starting consumer for new user")
			ctx, cancel := context.WithCancel(w.ctx)
			queueName := fmt.Sprintf("%s_incoming_messages", username)

			w.userConsumers[username] = &userConsumer{
				matrixUsername: username,
				cancel:         cancel,
			}

			w.wg.Add(1)
			go func(queue string, identityID uint, userCtx context.Context) {
				defer w.wg.Done()
				w.runUserConsumer(queue, identityID, userCtx)
			}(queueName, identityID, ctx)
		}
	}

	w.userConsumersMutex.Unlock()
}

func (w *WebhookWorker) getActiveMatrixIdentities() ([]models.MatrixIdentity, error) {
	var identities []models.MatrixIdentity
	err := w.db.DB().Find(&identities).Error
	return identities, err
}

func (w *WebhookWorker) runUserConsumer(queueName string, matrixIdentityID uint, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("Webhook consumer panic: %v\n%s", r, debug.Stack()))
		}
	}()

	retryDelay := 1 * time.Second
	maxRetryDelay := 60 * time.Second
	connectionSuccessThreshold := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		logger.Info("Webhook consumer: Connecting to queue")
		logger.Debug(fmt.Sprintf("Webhook consumer connecting to queue: %s", queueName))
		connectionStart := time.Now()
		err := w.runUserConsumerLoop(matrixIdentityID, queueName, ctx)
		connectionDuration := time.Since(connectionStart)

		if err == nil {
			return
		}

		if connectionDuration >= connectionSuccessThreshold {
			retryDelay = 1 * time.Second
		}

		select {
		case <-ctx.Done():
			return
		default:
			logger.Warn(fmt.Sprintf("Webhook consumer: Connection lost after %v, reconnecting in %v", connectionDuration.Round(time.Second), retryDelay))
			time.Sleep(retryDelay)
			retryDelay *= 2
			if retryDelay > maxRetryDelay {
				retryDelay = maxRetryDelay
			}
		}
	}
}

func (w *WebhookWorker) runUserConsumerLoop(matrixIdentityID uint, queueName string, parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	consumer, err := rabbitmq.NewConsumer(w.rabbitURL)
	if err != nil {
		return err
	}
	defer consumer.Close()

	deliveryHandler := func(delivery amqp.Delivery) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("Webhook consumer: Message handler panic: %v\n%s", r, debug.Stack()))
				delivery.Nack(false, false)
			}
		}()

		var msg IncomingMessage
		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			logger.Error(fmt.Sprintf("Webhook consumer: Message unmarshal failed: %v", err))
			delivery.Nack(false, false)
			return err
		}

		webhooks, err := models.FindActiveWebhooksByIdentity(w.db.DB(), matrixIdentityID)
		if err != nil {
			logger.Error(fmt.Sprintf("Webhook consumer: Failed to fetch webhooks: %v", err))
			delivery.Nack(false, true)
			return err
		}

		if len(webhooks) == 0 {
			logger.Debug("Webhook consumer: No active webhooks, skipping message")
			delivery.Ack(false)
			return nil
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			logger.Error(fmt.Sprintf("Webhook consumer: Failed to marshal message: %v", err))
			delivery.Nack(false, false)
			return err
		}

		var wg sync.WaitGroup
		for _, webhook := range webhooks {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				w.postToWebhook(url, jsonData)
			}(webhook.URL)
		}
		wg.Wait()

		logger.Info(fmt.Sprintf("Webhook consumer: Message delivered to %d webhook(s)", len(webhooks)))
		delivery.Ack(false)
		return nil
	}

	opts := rabbitmq.ManualAckOptions()
	opts.CreateQueue = true
	opts.CreateExchange = true
	opts.ExchangeType = "topic"
	opts.BindExchange = w.exchangeName
	opts.BindingKey = w.bindingKey

	err = consumer.Consume(ctx, queueName, deliveryHandler, cancel, opts)
	if err != nil {
		return err
	}

	logger.Info("Webhook consumer: Connected and listening")
	logger.Debug(fmt.Sprintf("Webhook consumer connected: queue='%s', exchange='%s', key='%s'", queueName, w.exchangeName, w.bindingKey))

	<-ctx.Done()

	select {
	case <-parentCtx.Done():
		return nil
	default:
		return fmt.Errorf("connection closed")
	}
}

func (w *WebhookWorker) postToWebhook(url string, data []byte) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		logger.Error(fmt.Sprintf("Webhook consumer: Failed to create HTTP request: %v", err))
		logger.Debug(fmt.Sprintf("Webhook request creation failed for URL: %s", url))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Shortmesh-Webhook/1.0")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Webhook consumer: HTTP POST failed: %v", err))
		logger.Debug(fmt.Sprintf("Webhook POST failed to: %s", url))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Info(fmt.Sprintf("Webhook consumer: Successfully delivered (status: %d)", resp.StatusCode))
		logger.Debug(fmt.Sprintf("Webhook delivered to %s (status: %d)", url, resp.StatusCode))
	} else {
		logger.Warn(fmt.Sprintf("Webhook consumer: Non-2xx response (status: %d)", resp.StatusCode))
		logger.Debug(fmt.Sprintf("Webhook %s returned status: %d", url, resp.StatusCode))
	}
}
