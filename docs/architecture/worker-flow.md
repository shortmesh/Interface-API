# Worker Flow

## Sequence Diagram

```mermaid
sequenceDiagram
    participant Main as Main Process
    participant Worker as Worker Instance
    participant RMQ as RabbitMQ
    participant Throttler
    participant MC as Matrix Client

    Main->>Main: Initialize configuration
    Main->>Throttler: Create shared throttler
    
    loop Spawn N workers
        Main->>Worker: Start worker
    end
    
    Worker->>MC: Initialize client
    Worker->>RMQ: Connect to message queue
    Note over Worker: Worker ready
    
    loop Process messages
        RMQ->>Worker: Deliver message
        Worker->>Throttler: Check rate limit
        
        alt Rate limited
            Throttler-->>Worker: Deny
            Worker->>RMQ: Delay message
        else Allowed
            Throttler-->>Worker: Allow
            Worker->>MC: Send message
            
            alt Success
                MC-->>Worker: Success
                Worker->>RMQ: Acknowledge
            else Failure
                MC-->>Worker: Error
                Worker->>RMQ: Discard
            end
        end
    end
    
    Note over Main: Shutdown signal
    Main->>Worker: Stop
    Worker-->>Main: Stopped
```


## Message Flow

1. **Initialization**: Workers connect to RabbitMQ and Matrix Client
2. **Consumption**: Workers receive messages from the queue
3. **Rate Limiting**: Throttler checks if message can be sent immediately
4. **Delivery**: Messages are forwarded to Matrix Client
5. **Acknowledgment**: Successful deliveries are acknowledged, failures are discarded
