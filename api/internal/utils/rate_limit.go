package utils

import (
	"sync"
	"time"
)

type rateWindow struct {
	start time.Time
	hits  int
}

type RateLimiter struct {
	limit     int
	window    time.Duration
	mutex     sync.Mutex
	windows   map[string]*rateWindow
	lastPurge time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, window: window, windows: map[string]*rateWindow{}, lastPurge: time.Now()}
}

// Allow records a hit for key and returns how long to wait when the limit is reached
func (limiter *RateLimiter) Allow(key string) (bool, time.Duration) {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	now := time.Now()
	if now.Sub(limiter.lastPurge) > limiter.window {
		for key, window := range limiter.windows {
			if now.Sub(window.start) >= limiter.window {
				delete(limiter.windows, key)
			}
		}
		limiter.lastPurge = now
	}

	window, exists := limiter.windows[key]
	if !exists || now.Sub(window.start) >= limiter.window {
		window = &rateWindow{start: now}
		limiter.windows[key] = window
	}
	if window.hits >= limiter.limit {
		return false, window.start.Add(limiter.window).Sub(now)
	}
	window.hits++
	return true, 0
}
