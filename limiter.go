package limiter

import (
	"math"
	"sync"
	"time"
)

// Limit defines the maximum frequency of some events.
// Limit is represented as a number of events per second.
type Limit uint64

// Infinite is the infinite rate limit; it allows all events.
// If you don't send the option WithMaxLimit, the limiter will use an infinite limit.
const Infinite Limit = math.MaxUint64

// If you don't send option the WithInterval, the limiter will use this interval.
const defaultInterval = time.Hour

// Limiter responsible for managing allows requests.
type Limiter struct {
	mu             sync.RWMutex
	maxLimitInTime Limit
	current        Limit
	interval       time.Duration
}

// WithMaxLimit set Limiter.maxLimitInTime.
func WithMaxLimit(limit Limit) func(*Limiter) {
	return func(l *Limiter) {
		l.maxLimitInTime = limit
	}
}

// WithInterval set Limiter.maxLimitInTime.
func WithInterval(interval time.Duration) func(*Limiter) {
	return func(l *Limiter) {
		l.interval = interval
	}
}

// New build and returns new instance Limiter.
func New(opts ...func(*Limiter)) *Limiter {
	l := &Limiter{
		mu:             sync.RWMutex{},
		maxLimitInTime: Infinite,
		current:        0,
		interval:       defaultInterval,
	}

	for i := range opts {
		opts[i](l)
	}

	return l
}

// Allow checks available for calling requests.
// If it calls in the first time
// will start goroutine to nullify the current limit after interval.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.current == 0 && l.maxLimitInTime != Infinite {
		go l.cleanupLimitAfterInterval()
	}

	var allow bool
	switch {
	case l.current < l.maxLimitInTime:
		allow = true
	}

	if allow {
		l.current++
	}

	return allow
}

func (l *Limiter) cleanupLimitAfterInterval() {
	select {
	case <-time.After(l.interval):
		l.mu.Lock()
		defer l.mu.Unlock()

		l.current = 0
	}
}
