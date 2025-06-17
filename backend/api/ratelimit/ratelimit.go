package ratelimit

import (
	"goaway/backend/logging"
	"sync"
	"time"
)

var log = logging.GetLogger()

type RateLimiterConfig struct {
	Enabled  bool `json:"-" yaml:"enabled"`
	MaxTries int  `json:"-" yaml:"maxTries"`
	Window   int  `json:"-" yaml:"window"`
}

type RateLimiter struct {
	Config   RateLimiterConfig
	mutex    *sync.RWMutex
	attempts map[string][]time.Time
}

func NewRateLimiter(enabled bool, maxTries int, window int) *RateLimiter {
	config := RateLimiterConfig{
		Enabled:  enabled,
		MaxTries: maxTries,
		Window:   window,
	}

	if !enabled {
		log.Warning("Rate limit is disabled")
	}

	return &RateLimiter{
		Config:   config,
		mutex:    &sync.RWMutex{},
		attempts: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) CheckLimit(identifier string) (bool, int) {
	if !rl.Config.Enabled {
		return true, 0
	}

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	window := time.Duration(rl.Config.Window) * time.Minute
	now := time.Now()
	cutoff := now.Add(-window)
	attempts := rl.attempts[identifier]
	validAttempts := make([]time.Time, 0)

	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			validAttempts = append(validAttempts, attempt)
		}
	}

	if len(validAttempts) >= rl.Config.MaxTries {
		oldestAttempt := validAttempts[0]
		timeUntilReset := window - now.Sub(oldestAttempt)

		log.Info(
			"Rate limit exceeded for %s: %d/%d attempts, reset in %s\n",
			identifier,
			len(validAttempts),
			rl.Config.MaxTries,
			timeUntilReset,
		)

		return false, int(timeUntilReset.Seconds())
	}

	validAttempts = append(validAttempts, now)
	rl.attempts[identifier] = validAttempts

	return true, 0
}
