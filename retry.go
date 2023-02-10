package gojob

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	DefaultMaxRetries   = 5
	DefaultInitialDelay = 100 * time.Millisecond
	DefaultMaxDelay     = 100 * time.Millisecond
)

func Run[T any](ctx context.Context, fn func(context.Context) (T, error), maxTries int, initialDelay, maxDelay time.Duration, exitErrs []error) (T, error) {
	maxTries = orElse(maxTries > 0, maxTries, DefaultMaxRetries)
	initialDelay = orElse(initialDelay > 0, initialDelay, DefaultInitialDelay)
	maxDelay = orElse(maxDelay > 0, maxDelay, DefaultMaxDelay)

	attempts := 0
	for {
		res, err := fn(ctx)
		if err == nil || include(exitErrs, err) {
			return res, err
		}

		attempts++
		if attempts == maxTries {
			return res, fmt.Errorf("max attempt exceeded: %w", err)
		}

		ticker := time.NewTimer(nextBackOff(attempts, initialDelay, maxDelay))
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return res, fmt.Errorf("context timeout: %w", err)
		}
	}
}

func nextBackOff(attempt int, initialDelay, maxDelay time.Duration) time.Duration {
	max := float64(maxDelay)
	min := float64(initialDelay)

	randPool := sync.Pool{
		New: func() any {
			return rand.Float64()
		},
	}

	d := min * math.Pow(2, float64(attempt))
	d = orElse(d < max, d, max)

	rand := randPool.Get().(float64)
	return time.Duration(rand*(d-min) + min)
}
