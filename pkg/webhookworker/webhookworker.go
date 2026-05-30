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
	ID        string `json:"ID"`
	IsContact bool   `json:"IsContact"`
	Type      string `json:"Type"`
	From      string `json:"From"`
	To        string `json:"To"`
	Message   string `json:"Message"`
	DeviceID  string `json:"DeviceId"`
	Media     Media  `json:"Media"`
}

type WebhookMediaInfo struct {
	Size     float64 `json:"size"`
	MimeType string  `json:"mime_type"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	BlurHash string  `json:"blur_hash"`
}

type WebhookMedia struct {
	Content []byte           `json:"content"`
	Info    WebhookMediaInfo `json:"info"`
}

type WebhookMessage struct {
	ID        string       `json:"id"`
	IsContact bool         `json:"is_contact"`
	Type      string       `json:"type"`
	From      string       `json:"from"`
	To        string       `json:"to"`
	Message   string       `json:"message"`
	DeviceID  string       `json:"device_id"`
	Media     WebhookMedia `json:"media"`
}

func (m *IncomingMessage) ToWebhookMessage() *WebhookMessage {
	return &WebhookMessage{
		ID:        m.ID,
		IsContact: m.IsContact,
		Type:      m.Type,
		From:      m.From,
		To:        m.To,
		Message:   m.Message,
		DeviceID:  m.DeviceID,
		Media: WebhookMedia{
			Content: m.Media.Content,
			Info: WebhookMediaInfo{
				Size:     m.Media.Info.Size,
				MimeType: m.Media.Info.MimeType,
				Width:    m.Media.Info.Width,
				Height:   m.Media.Info.Height,
				BlurHash: m.Media.Info.BlurHash,
			},
		},
	}
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
	queueSuffix        string
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

	queueSuffix := os.Getenv("WEBHOOK_INCOMING_QUEUE_SUFFIX")
	if queueSuffix == "" {
		queueSuffix = "_incoming_messages"
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
		queueSuffix:     queueSuffix,
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

	usernameIdentities := make(map[string][]uint)
	for _, identity := range identities {
		usernameIdentities[identity.MatrixUsername] = append(
			usernameIdentities[identity.MatrixUsername],
			identity.ID,
		)
	}

	w.userConsumersMutex.Lock()

	for username, consumer := range w.userConsumers {
		if _, exists := usernameIdentities[username]; !exists {
			logger.Info("Stopping consumer for removed user")
			consumer.cancel()
			delete(w.userConsumers, username)
		}
	}

	for username, identityIDs := range usernameIdentities {
		if _, exists := w.userConsumers[username]; !exists {
			logger.Info("Starting consumer for new user")
			ctx, cancel := context.WithCancel(w.ctx)
			queueName := username + w.queueSuffix

			w.userConsumers[username] = &userConsumer{
				matrixUsername: username,
				cancel:         cancel,
			}

			w.wg.Add(1)
			go func(queue string, identityIDs []uint, userCtx context.Context) {
				defer w.wg.Done()
				w.runUserConsumer(queue, identityIDs, userCtx)
			}(queueName, identityIDs, ctx)
		}
	}

	w.userConsumersMutex.Unlock()
}

func (w *WebhookWorker) getActiveMatrixIdentities() ([]models.MatrixIdentity, error) {
	var identities []models.MatrixIdentity
	err := w.db.DB().Find(&identities).Error
	return identities, err
}

func (w *WebhookWorker) runUserConsumer(queueName string, matrixIdentityIDs []uint, ctx context.Context) {
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
		err := w.runUserConsumerLoop(matrixIdentityIDs, queueName, ctx)
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

func (w *WebhookWorker) runUserConsumerLoop(matrixIdentityIDs []uint, queueName string, parentCtx context.Context) error {
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

		var allWebhooks []models.Webhook
		seenURLs := make(map[string]bool)

		for _, identityID := range matrixIdentityIDs {
			webhooks, err := models.FindActiveWebhooksByIdentity(w.db.DB(), identityID)
			if err != nil {
				logger.Error(fmt.Sprintf("Webhook consumer: Failed to fetch webhooks for identity %d: %v", identityID, err))
				continue
			}

			for _, webhook := range webhooks {
				if !seenURLs[webhook.URL] {
					allWebhooks = append(allWebhooks, webhook)
					seenURLs[webhook.URL] = true
				}
			}
		}

		if len(allWebhooks) == 0 {
			logger.Debug("Webhook consumer: No active webhooks, skipping message")
			delivery.Ack(false)
			return nil
		}

		webhookMsg := msg.ToWebhookMessage()
		jsonData, err := json.Marshal(webhookMsg)
		if err != nil {
			logger.Error(fmt.Sprintf("Webhook consumer: Failed to marshal message: %v", err))
			delivery.Nack(false, false)
			return err
		}

		var wg sync.WaitGroup
		for _, webhook := range allWebhooks {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				w.postToWebhook(url, jsonData)
			}(webhook.URL)
		}
		wg.Wait()

		logger.Info(fmt.Sprintf("Webhook consumer: Message delivered to %d webhook(s)", len(allWebhooks)))
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

func (w *WebhookWorker) postToWebhook(url string, data []byte) (statusCode int, shouldRetry bool) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		logger.Error(fmt.Sprintf("Webhook consumer: Failed to create HTTP request: %v", err))
		logger.Debug(fmt.Sprintf("Webhook request creation failed for URL: %s", url))
		return 0, true
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Shortmesh-Webhook/1.0")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Webhook consumer: HTTP POST failed: %v", err))
		logger.Debug(fmt.Sprintf("Webhook POST failed to: %s", url))
		return 0, true
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode
	shouldRetry = statusCode >= 500

	if statusCode >= 200 && statusCode < 300 {
		logger.Info(fmt.Sprintf("Webhook consumer: Successfully delivered (status: %d)", statusCode))
		logger.Debug(fmt.Sprintf("Webhook delivered to %s (status: %d)", url, statusCode))
	} else if statusCode >= 500 {
		logger.Warn(fmt.Sprintf("Webhook consumer: Server error, will retry (status: %d)", statusCode))
		logger.Debug(fmt.Sprintf("Webhook %s returned 5xx status: %d", url, statusCode))
	} else {
		logger.Warn(fmt.Sprintf("Webhook consumer: Client error, will not retry (status: %d)", statusCode))
		logger.Debug(fmt.Sprintf("Webhook %s returned non-retryable status: %d", url, statusCode))
	}

	return statusCode, shouldRetry
}
