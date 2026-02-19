package rabbitmq

import (
	"context"
	"fmt"
)

type Consumer struct {
	*Client
}

func NewConsumer(url string) (*Consumer, error) {
	client, err := dial(url)
	if err != nil {
		return nil, err
	}

	return &Consumer{Client: client}, nil
}

func (c *Consumer) Consume(
	ctx context.Context,
	queueName string,
	handler DeliveryHandler,
	cancelFunc context.CancelFunc,
	opts ConsumeOptions,
) error {
	if opts.CreateExchange && opts.BindExchange != "" {
		exchangeType := opts.ExchangeType
		if exchangeType == "" {
			exchangeType = "direct"
		}
		if err := c.declareExchange(DefaultExchangeConfig(opts.BindExchange, exchangeType)); err != nil {
			return fmt.Errorf("failed to declare exchange '%s': %w", opts.BindExchange, err)
		}
	}

	if opts.CreateQueue {
		if err := c.declareQueue(DefaultQueueConfig(queueName)); err != nil {
			return fmt.Errorf("failed to declare queue '%s': %w", queueName, err)
		}
	}

	if opts.BindExchange != "" {
		if err := c.bindQueue(queueName, opts.BindExchange, opts.BindingKey); err != nil {
			return fmt.Errorf("failed to bind queue '%s' to exchange '%s': %w", queueName, opts.BindExchange, err)
		}
	}

	if err := c.setQos(opts.PrefetchCount, opts.PrefetchSize, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	config := consumeConfig{
		queueName: queueName,
		consumer:  queueName,
		autoAck:   opts.AutoAck,
		exclusive: opts.Exclusive,
		noLocal:   opts.NoLocal,
		noWait:    opts.NoWait,
		args:      opts.Args,
	}

	msgs, err := c.consume(config)
	if err != nil {
		return fmt.Errorf("failed to start consuming from queue '%s': %w", queueName, err)
	}

	startMessageLoop(ctx, msgs, handler, c.channel, cancelFunc)
	return nil
}

func (c *Consumer) DeclareQueue(config QueueConfig) error {
	return c.declareQueue(config)
}

func (c *Consumer) DoesQueueExist(queueName string) (bool, error) {
	return c.queueExists(queueName)
}

func (c *Consumer) Close() error {
	return c.close()
}
