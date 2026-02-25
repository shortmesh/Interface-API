# Throttler

A rate limiter implementation using the Token Bucket Algorithm with configurable jitter for randomized timing intervals.

## Token Bucket Algorithm

The token bucket algorithm maintains a virtual bucket that accumulates tokens at a constant rate. Each request consumes one token. If no tokens are available, the request is denied. The bucket starts empty and refills continuously at a rate defined by `Rate / Interval`. Jitter randomizes the effective refill rate within configured bounds.

## Configuration Structure

```go
type PlatformConfig struct {
    Rate      int           // Number of tokens per interval
    Interval  time.Duration // Time window for token refill
    JitterMin float64       // Minimum timing multiplier
    JitterMax float64       // Maximum timing multiplier
}
```

Each platform maintains independent throttling configuration. The rate defines how many tokens accumulate per interval, while jitter bounds randomize the actual timing.

## Default Platform Configurations

| Platform  | Rate | Interval | JitterMin | JitterMax | Effective Range |
|-----------|------|----------|-----------|-----------|-----------------|
| Platform1 | 1    | 30s      | 0.8       | 1.3       | 24s - 39s       |
| Platform2 | 1    | 20s      | 0.8       | 1.3       | 16s - 26s       |
| Platform3 | 1    | 40s      | 0.8       | 1.3       | 32s - 52s       |

## Example 1: Low Rate Configuration

Configuration with 1 message per 30 seconds. Jitter range 0.8-1.3 produces variable intervals.

**Refill Rate**: 1 token / 30s = 0.0333 tokens/second  
**Effective Range**: 24s to 39s

| Time     | Tokens | Action | Jitter Applied |
|----------|--------|--------|----------------|
| 14:00:00 | 0.00   | DENY   | Bucket empty   |
| 14:00:27 | 1.00   | ALLOW  | 27s (0.9×30s)  |
| 14:00:27 | 0.00   | DENY   | Token consumed |
| 14:01:02 | 1.00   | ALLOW  | 35s (1.17×30s) |
| 14:01:34 | 1.00   | ALLOW  | 32s (1.07×30s) |

## Example 2: Medium Rate Configuration

Configuration with 1 message per 20 seconds. Faster refill rate with same jitter bounds.

**Refill Rate**: 1 token / 20s = 0.05 tokens/second  
**Effective Range**: 16s to 26s

| Time     | Tokens | Action | Jitter Applied |
|----------|--------|--------|----------------|
| 14:00:00 | 0.00   | DENY   | Bucket empty   |
| 14:00:18 | 1.00   | ALLOW  | 18s (0.9×20s)  |
| 14:00:18 | 0.00   | DENY   | Token consumed |
| 14:00:40 | 1.00   | ALLOW  | 22s (1.1×20s)  |
| 14:01:06 | 1.00   | ALLOW  | 26s (1.3×20s)  |

## Example 3: High Rate Configuration

Configuration with 5 messages per 30 seconds. Allows burst behavior once tokens accumulate.

**Refill Rate**: 5 tokens / 30s = 0.1667 tokens/second  
**Token Capacity**: 5  
**Per-Token Refill**: 6 seconds

| Time     | Tokens | Action             | Notes                 |
|----------|--------|--------------------|-----------------------|
| 14:00:00 | 0.00   | DENY               | Bucket empty          |
| 14:00:30 | 5.00   | Bucket Full        | All tokens available  |
| 14:00:30 | 4.00   | ALLOW (message 1)  | Token consumed        |
| 14:00:30 | 3.00   | ALLOW (message 2)  | Token consumed        |
| 14:00:30 | 2.00   | ALLOW (message 3)  | Token consumed        |
| 14:00:30 | 1.00   | ALLOW (message 4)  | Token consumed        |
| 14:00:30 | 0.00   | ALLOW (message 5)  | Last token consumed   |
| 14:00:30 | 0.00   | DENY               | No tokens available   |
| 14:00:36 | 1.00   | ALLOW              | 1 token refilled      |
| 14:00:42 | 1.00   | ALLOW              | 1 token refilled      |

## Multi-Tenant Isolation

The throttler maintains separate token buckets for each unique combination of platform and username. Each bucket operates independently with its own token count and refill schedule.

**Isolation Key Format**: `platform:username`

| Bucket Key      | Tokens | Independent State |
|-----------------|--------|-------------------|
| platform1:user1 | 4/10   | Separate bucket   |
| platform2:user1 | 6/10   | Separate bucket   |
| platform1:user2 | 2/10   | Separate bucket   |

## Implementation Usage

```go
// Initialize with default configurations
throttler := throttler.New()

// Check if request is allowed
if throttler.Allow("platform1", "user@example.org") {
    // Process request
    sendMessage()
} else {
    // Calculate wait time
    wait := throttler.WaitTime("platform1", "user@example.org")
    log.Printf("Rate limited. Retry in %v", wait)
}
```

## Jitter Calculation

Jitter randomizes the effective interval by applying a random multiplier between configured bounds. The formula generates a uniformly distributed random value within the jitter range.

**Formula**:

```
jitter = jitterMin + random() * (jitterMax - jitterMin)
effectiveInterval = baseInterval * jitter
```

**Example with base interval 30s, jitter bounds 0.8-1.3**:

| random() | jitter | Calculation       | Effective Interval |
|----------|--------|-------------------|--------------------|
| 0.0      | 0.80   | 30s × 0.80        | 24.0s              |
| 0.2      | 0.90   | 30s × 0.90        | 27.0s              |
| 0.4      | 1.00   | 30s × 1.00        | 30.0s              |
| 0.7      | 1.15   | 30s × 1.15        | 34.5s              |
| 1.0      | 1.30   | 30s × 1.30        | 39.0s              |

## Thread Safety

The throttler uses concurrent-safe data structures to handle multiple simultaneous requests across goroutines.

| Component    | Mechanism            | Purpose                          |
|--------------|----------------------|----------------------------------|
| sync.Map     | Concurrent map       | Thread-safe limiter storage      |
| Mutex        | Per-bucket lock      | Atomic token operations          |
| LoadOrStore  | Atomic operation     | Race-free limiter initialization |

Each token bucket is created lazily on first access for a platform-username pair and reused for subsequent requests.
