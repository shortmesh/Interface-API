package throttler

import (
	"math/rand"
	"sync"
	"time"
)

type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
	jitterMin  float64
	jitterMax  float64
	mu         sync.Mutex
}

func newTokenBucket(rate int, interval time.Duration, jitterMin, jitterMax float64) *tokenBucket {
	if jitterMin == 0 {
		jitterMin = 1.0
	}
	if jitterMax == 0 {
		jitterMax = 1.0
	}
	refillRate := float64(rate) / interval.Seconds()
	return &tokenBucket{
		tokens:     0,
		maxTokens:  float64(rate),
		refillRate: refillRate,
		lastRefill: time.Now(),
		jitterMin:  jitterMin,
		jitterMax:  jitterMax,
	}
}

func (tb *tokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	jitter := tb.jitterMin + rand.Float64()*(tb.jitterMax-tb.jitterMin)
	adjustedRefillRate := tb.refillRate / jitter

	tb.tokens += elapsed * adjustedRefillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefill = now

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

func (tb *tokenBucket) WaitTime() time.Duration {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.tokens >= 1.0 {
		return 0
	}

	jitter := tb.jitterMin + rand.Float64()*(tb.jitterMax-tb.jitterMin)
	tokensNeeded := 1.0 - tb.tokens
	waitSeconds := (tokensNeeded / tb.refillRate) * jitter
	return time.Duration(waitSeconds * float64(time.Second))
}
