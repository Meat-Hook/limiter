package limiter

import (
	"context"
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
	mu       sync.RWMutex
	limit    Limit
	current  Limit
	interval time.Duration

	gradualRecovery bool

	done chan struct{}
}

// New build and returns new instance Limiter.
func New(opts ...func(*Limiter)) *Limiter {
	l := &Limiter{
		mu:              sync.RWMutex{},
		limit:           Infinite,
		current:         0,
		interval:        defaultInterval,
		gradualRecovery: false,
		done:            make(chan struct{}, 1),
	}

	for i := range opts {
		opts[i](l)
	}

	go l.cleanupLimitAfterInterval()

	return l
}

// Allow checks available for calling requests.
// If it calls in the first time
// will start goroutine to nullify the current limit after interval.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	var allow bool
	switch {
	case l.current < l.limit:
		allow = true
	}

	if allow {
		l.current++
	}

	return allow
}

// Current returns current limit.
func (l *Limiter) Current() Limit {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.current
}

// Wait waits when we can call Allow.
// If ctx done, will return false.
func (l *Limiter) Wait(ctx context.Context) bool {
	allow := make(chan struct{}, 1)
	go func() {
		defer close(allow)

		for {
			l.mu.RLock()
			if l.current < l.limit {
				allow <- struct{}{}
				break
			}
			l.mu.RUnlock()
		}
	}()

	select {
	case <-ctx.Done():
		return false
	case <-allow:
		return l.Allow()
	case <-l.done:
		return false
	}
}

// Close limiter workers.
// It must be call.
func (l *Limiter) Close() {
	close(l.done)
}

func (l *Limiter) cleanupLimitAfterInterval() {
	l.mu.RLock()
	interval := l.interval
	repair := l.limit
	l.mu.RUnlock()

	if l.gradualRecovery {
		x := int(interval.Seconds()) / int(repair)
		interval = time.Second * time.Duration(x)
		repair = 1
	}

	for {
		select {
		case <-time.After(l.interval):
			l.mu.Lock()

			if l.gradualRecovery {
				l.current -= repair
				if l.current < 0 {
					l.current = 0
				}
			} else {
				l.current = 0
			}

			l.mu.Unlock()
		case <-l.done:
			return
		}
	}
}
