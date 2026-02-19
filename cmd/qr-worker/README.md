# RabbitMQ QR Worker

Consumes messages from a RabbitMQ queue and renders each message as an ASCII QR code in the terminal.

## Execution

```bash
# Default configuration
./bin/qr-worker

# Custom connection and queue
./bin/qr-worker -url amqp://user:pass@rabbitmq:5672/ -queue myqueue
```

## Flags

| Flag     | Description                | Default                              |
| -------- | -------------------------- | ------------------------------------ |
| `-url`   | AMQP connection URL        | `amqp://guest:guest@localhost:5672/` |
| `-queue` | Queue name to consume from | `default`                            |

## Example

```bash
./bin/qr-worker -url amqp://guest:guest@localhost:5672/ -queue orders
```

## Behavior

* Establishes AMQP connection
* Subscribes to the specified queue
* For each message:
  * Generates and prints an ASCII QR code
  * Prints the raw message payload
* Runs until terminated (`Ctrl+C`)
