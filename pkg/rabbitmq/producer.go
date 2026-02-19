package rabbitmq

import "fmt"

type Producer struct {
	*Client
}

func NewProducer(url string) (*Producer, error) {
	client, err := dial(url)
	if err != nil {
		return nil, err
	}

	return &Producer{Client: client}, nil
}

func (p *Producer) Publish(exchange, routingKey string, message any, opts PublishOptions) error {
	body, err := marshalMessage(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := p.publish(exchange, routingKey, body, opts); err != nil {
		return fmt.Errorf("failed to publish message to exchange '%s': %w", exchange, err)
	}
	return nil
}

func (p *Producer) DeclareExchange(exchangeName, exchangeType string) error {
	if err := p.declareExchange(DefaultExchangeConfig(exchangeName, exchangeType)); err != nil {
		return fmt.Errorf("failed to declare exchange '%s': %w", exchangeName, err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.close()
}
