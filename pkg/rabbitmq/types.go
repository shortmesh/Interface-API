package rabbitmq

import "github.com/streadway/amqp"

type DeliveryHandler func(amqp.Delivery) error

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type ExchangeConfig struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

type ConsumeOptions struct {
	AutoAck        bool
	CreateQueue    bool
	CreateExchange bool
	ExchangeType   string
	PrefetchCount  int
	PrefetchSize   int
	Exclusive      bool
	NoLocal        bool
	NoWait         bool
	Args           amqp.Table
	BindExchange   string
	BindingKey     string
}

type PublishOptions struct {
	ContentType  string
	DeliveryMode uint8
	Priority     uint8
	Expiration   string
	Mandatory    bool
	Immediate    bool
	Headers      amqp.Table
}

func DefaultQueueConfig(name string) QueueConfig {
	return QueueConfig{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
}

func DefaultExchangeConfig(name, exchangeType string) ExchangeConfig {
	return ExchangeConfig{
		Name:       name,
		Type:       exchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
}

func DefaultConsumeOptions() ConsumeOptions {
	return ConsumeOptions{
		AutoAck:        true,
		CreateQueue:    false,
		CreateExchange: false,
		ExchangeType:   "",
		PrefetchCount:  10,
		PrefetchSize:   0,
		Exclusive:      false,
		NoLocal:        false,
		NoWait:         false,
		Args:           nil,
		BindExchange:   "",
		BindingKey:     "",
	}
}

func ManualAckOptions() ConsumeOptions {
	return ConsumeOptions{
		AutoAck:        false,
		CreateQueue:    false,
		CreateExchange: false,
		ExchangeType:   "",
		PrefetchCount:  1,
		PrefetchSize:   0,
		Exclusive:      false,
		NoLocal:        false,
		NoWait:         false,
		Args:           nil,
		BindExchange:   "",
		BindingKey:     "",
	}
}

func DefaultPublishOptions() PublishOptions {
	return PublishOptions{
		ContentType:  "application/json",
		DeliveryMode: 2, // Persistent
		Priority:     0,
		Expiration:   "",
		Mandatory:    false,
		Immediate:    false,
		Headers:      nil,
	}
}

func TransientPublishOptions() PublishOptions {
	return PublishOptions{
		ContentType:  "application/json",
		DeliveryMode: 1, // Non-persistent
		Priority:     0,
		Expiration:   "",
		Mandatory:    false,
		Immediate:    false,
		Headers:      nil,
	}
}
