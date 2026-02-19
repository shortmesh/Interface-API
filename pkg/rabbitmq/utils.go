package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type consumeConfig struct {
	queueName string
	consumer  string
	autoAck   bool
	exclusive bool
	noLocal   bool
	noWait    bool
	args      amqp.Table
}

func dial(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *Client) close() error {
	var err error
	if c.channel != nil {
		if closeErr := c.channel.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if c.conn != nil {
		if closeErr := c.conn.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
			}
		}
	}
	return err
}

func (c *Client) declareQueue(config QueueConfig) error {
	_, err := c.channel.QueueDeclare(
		config.Name,
		config.Durable,
		config.AutoDelete,
		config.Exclusive,
		config.NoWait,
		config.Args,
	)
	return err
}

func (c *Client) declareExchange(config ExchangeConfig) error {
	return c.channel.ExchangeDeclare(
		config.Name,
		config.Type,
		config.Durable,
		config.AutoDelete,
		config.Internal,
		config.NoWait,
		config.Args,
	)
}

func (c *Client) bindQueue(queueName, exchangeName, routingKey string) error {
	return c.channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
}

func (c *Client) queueExists(queueName string) (bool, error) {
	_, err := c.channel.QueueInspect(queueName)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) setQos(prefetchCount, prefetchSize int, global bool) error {
	return c.channel.Qos(prefetchCount, prefetchSize, global)
}

func (c *Client) consume(config consumeConfig) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		config.queueName,
		config.consumer,
		config.autoAck,
		config.exclusive,
		config.noLocal,
		config.noWait,
		config.args,
	)
}

func (c *Client) publish(exchange, routingKey string, body []byte, opts PublishOptions) error {
	return c.channel.Publish(
		exchange,
		routingKey,
		opts.Mandatory,
		opts.Immediate,
		amqp.Publishing{
			ContentType:  opts.ContentType,
			DeliveryMode: opts.DeliveryMode,
			Priority:     opts.Priority,
			Expiration:   opts.Expiration,
			Headers:      opts.Headers,
			Body:         body,
		},
	)
}

func marshalMessage(message any) ([]byte, error) {
	return json.Marshal(message)
}

func startMessageLoop(
	ctx context.Context,
	msgs <-chan amqp.Delivery,
	handler func(amqp.Delivery) error,
	channel *amqp.Channel,
	cancelFunc context.CancelFunc,
) {
	closeChan := make(chan *amqp.Error)
	channel.NotifyClose(closeChan)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-closeChan:
				if err != nil {
					fmt.Printf("RabbitMQ channel closed: %v\n", err)
				}
				if cancelFunc != nil {
					cancelFunc()
				}
				return
			case msg, ok := <-msgs:
				if !ok {
					fmt.Println("Consumer channel closed, queue may have been deleted")
					if cancelFunc != nil {
						cancelFunc()
					}
					return
				}
				if err := handler(msg); err != nil {
					continue
				}
			}
		}
	}()
}
