package rabbitmq

import (
	"context"
	"fmt"

	"github.com/streadway/amqp"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

type MessageHandler func([]byte) error

func NewConsumer(url string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		url:     url,
	}, nil
}

func (c *Consumer) ConsumeQueue(ctx context.Context, queueName string, handler MessageHandler) error {
	ok, err := c.DoesQueueExist(queueName)
	if !ok {
		return fmt.Errorf("queue '%s' does not exist: %w", queueName, err)
	}

	msgs, err := c.channel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				if err := handler(msg.Body); err != nil {
					continue
				}
			}
		}
	}()

	return nil
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Consumer) DoesQueueExist(queueName string) (bool, error) {
	_, err := c.channel.QueueInspect(queueName)
	if err != nil {
		return false, err
	}
	return true, nil
}
