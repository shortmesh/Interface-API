package throttler

import "time"

type PlatformConfig struct {
	Rate      int           // Messages per interval
	Interval  time.Duration // Time window
	JitterMin float64       // Minimum jitter multiplier (e.g., 0.8 = 80% of interval)
	JitterMax float64       // Maximum jitter multiplier (e.g., 1.2 = 120% of interval)
}

var DefaultPlatformConfigs = map[string]PlatformConfig{
	"default": {
		Rate:      1,
		Interval:  8 * time.Second,
		JitterMin: 0.75,
		JitterMax: 1.25,
	},
}
