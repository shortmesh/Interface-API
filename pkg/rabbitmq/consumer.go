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

func (c *Consumer) ConsumeQueue(ctx context.Context, queueName string, handler MessageHandler, cancelFunc context.CancelFunc) error {
	ok, err := c.DoesQueueExist(queueName)
	if !ok {
		return fmt.Errorf("queue '%s' does not exist: %w", queueName, err)
	}

	err = c.channel.Qos(
		10,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		queueName,
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	closeChan := make(chan *amqp.Error)
	c.channel.NotifyClose(closeChan)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-closeChan:
				if err != nil {
					fmt.Printf("Channel closed: %v\n", err)
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
				if err := handler(msg.Body); err != nil {
					continue
				}
			}
		}
	}()

	return nil
}

func (c *Consumer) Close() error {
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

func (c *Consumer) DoesQueueExist(queueName string) (bool, error) {
	_, err := c.channel.QueueInspect(queueName)
	if err != nil {
		return false, err
	}
	return true, nil
}
