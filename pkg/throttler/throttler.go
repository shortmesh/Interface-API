package throttler

import (
	"fmt"
	"sync"
	"time"
)

type Throttler struct {
	limiters sync.Map
	configs  map[string]PlatformConfig
}

func New() *Throttler {
	return NewWithConfigs(DefaultPlatformConfigs)
}

func NewWithConfigs(configs map[string]PlatformConfig) *Throttler {
	return &Throttler{
		configs: configs,
	}
}

func (t *Throttler) getOrCreateLimiter(platform, username string) *tokenBucket {
	key := fmt.Sprintf("%s:%s", platform, username)

	if limiter, ok := t.limiters.Load(key); ok {
		return limiter.(*tokenBucket)
	}

	config, exists := t.configs[platform]
	if !exists {
		config = t.configs["default"]
	}

	limiter := newTokenBucket(
		config.Rate, config.Interval, config.JitterMin, config.JitterMax,
	)
	actual, loaded := t.limiters.LoadOrStore(key, limiter)
	if loaded {
		return actual.(*tokenBucket)
	}

	return limiter
}

func (t *Throttler) Allow(platform, username string) bool {
	limiter := t.getOrCreateLimiter(platform, username)
	return limiter.Allow()
}

func (t *Throttler) WaitTime(platform, username string) time.Duration {
	limiter := t.getOrCreateLimiter(platform, username)
	return limiter.WaitTime()
}
