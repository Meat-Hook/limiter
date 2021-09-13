package limiter

import "time"

// WithMaxLimit set Limiter.limit.
func WithMaxLimit(limit Limit) func(*Limiter) {
	return func(l *Limiter) {
		l.limit = limit
	}
}

// WithInterval set Limiter.interval.
func WithInterval(interval time.Duration) func(*Limiter) {
	return func(l *Limiter) {
		l.interval = interval
	}
}

// WithGradualRecovery set Limiter.gradualRecovery and Limiter.recoveryRate.
func WithGradualRecovery() func(*Limiter) {
	return func(l *Limiter) {
		l.gradualRecovery = true
	}
}
