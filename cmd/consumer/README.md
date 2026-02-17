# RabbitMQ Consumer with QR Code Display

This consumer connects to RabbitMQ, consumes messages from a queue, and displays each message as a QR code in the terminal.

## Usage

```bash
# Run with default settings (localhost RabbitMQ, 'default' queue)
./bin/consumer

# Specify custom RabbitMQ URL and queue
./bin/consumer -url amqp://user:pass@rabbitmq:5672/ -queue myqueue
```

## Flags

- `-url`: RabbitMQ connection URL (default: `amqp://guest:guest@localhost:5672/`)
- `-queue`: Queue name to consume from (default: `default`)

## Example

```bash
./bin/consumer -queue orders -url amqp://guest:guest@localhost:5672/
```

The consumer will:

1. Connect to RabbitMQ
2. Consume from the specified queue
3. Display each incoming message as a QR code in ASCII art
4. Print the message content below the QR code
5. Continue until interrupted with Ctrl+C
