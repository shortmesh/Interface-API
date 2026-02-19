package throttler

import (
	"testing"
	"time"
)

func TestThrottler_Allow(t *testing.T) {
	config := map[string]PlatformConfig{
		"test": {
			Rate:     2,
			Interval: 1 * time.Second,
		},
	}

	throttler := NewWithConfigs(config)
	platform := "test"
	username := "user1"

	if throttler.Allow(platform, username) {
		t.Error("First request should be throttled")
	}

	time.Sleep(500 * time.Millisecond)

	if !throttler.Allow(platform, username) {
		t.Error("Request after 0.5s should be allowed")
	}

	if throttler.Allow(platform, username) {
		t.Error("Second immediate request should be throttled")
	}

	time.Sleep(500 * time.Millisecond)

	if !throttler.Allow(platform, username) {
		t.Error("Request after another 0.5s should be allowed")
	}
}

func TestThrottler_IndependentPlatforms(t *testing.T) {
	config := map[string]PlatformConfig{
		"whatsapp": {
			Rate:      1,
			Interval:  1 * time.Second,
			JitterMin: 1.0,
			JitterMax: 1.0,
		},
		"telegram": {
			Rate:      1,
			Interval:  1 * time.Second,
			JitterMin: 1.0,
			JitterMax: 1.0,
		},
	}

	throttler := NewWithConfigs(config)
	username := "user1"

	throttler.Allow("whatsapp", username)
	throttler.Allow("telegram", username)

	time.Sleep(1200 * time.Millisecond)

	if !throttler.Allow("whatsapp", username) {
		t.Error("WhatsApp request should be allowed")
	}

	if !throttler.Allow("telegram", username) {
		t.Error("Telegram request should be allowed")
	}

	if throttler.Allow("whatsapp", username) {
		t.Error("Second WhatsApp request should be throttled")
	}

	if throttler.Allow("telegram", username) {
		t.Error("Second Telegram request should be throttled")
	}
}

func TestThrottler_IndependentUsers(t *testing.T) {
	config := map[string]PlatformConfig{
		"whatsapp": {
			Rate:      1,
			Interval:  1 * time.Second,
			JitterMin: 1.0,
			JitterMax: 1.0,
		},
	}

	throttler := NewWithConfigs(config)
	platform := "whatsapp"
	username1 := "user1"
	username2 := "user2"

	throttler.Allow(platform, username1)
	throttler.Allow(platform, username2)

	time.Sleep(1200 * time.Millisecond)

	if !throttler.Allow(platform, username1) {
		t.Error("User1 request should be allowed")
	}

	if !throttler.Allow(platform, username2) {
		t.Error("User2 request should be allowed")
	}

	if throttler.Allow(platform, username1) {
		t.Error("Second User1 request should be throttled")
	}

	if throttler.Allow(platform, username2) {
		t.Error("Second User2 request should be throttled")
	}
}

func TestThrottler_WaitTime(t *testing.T) {
	config := map[string]PlatformConfig{
		"test": {
			Rate:     1,
			Interval: 2 * time.Second,
		},
	}

	throttler := NewWithConfigs(config)
	platform := "test"
	username := "user1"

	throttler.Allow(platform, username)

	waitTime := throttler.WaitTime(platform, username)
	if waitTime <= 0 || waitTime > 2*time.Second {
		t.Errorf("Wait time should be between 0 and 2 seconds, got %v", waitTime)
	}
}

func TestThrottler_DefaultPlatform(t *testing.T) {
	config := map[string]PlatformConfig{
		"default": {
			Rate:      1,
			Interval:  1 * time.Second,
			JitterMin: 1.0,
			JitterMax: 1.0,
		},
	}
	username := "user1"

	throttler := NewWithConfigs(config)

	throttler.Allow("unknown_platform", username)

	time.Sleep(1200 * time.Millisecond)

	if !throttler.Allow("unknown_platform", username) {
		t.Error("Request after refill should use default config")
	}
}
