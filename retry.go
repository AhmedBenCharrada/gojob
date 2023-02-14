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
	defaultMaxRetries   = 5
	defaultInitialDelay = 100 * time.Millisecond
	defaultMaxDelay     = 100 * time.Millisecond
)

type parameters struct {
	maxTries               int
	initialDelay, maxDelay time.Duration
	exitErrs               []error
}

type parameter func(parameters) parameters

func WithMaxTries(tries int) parameter {
	return func(p parameters) parameters {
		p.maxTries = tries
		return p
	}
}

func WithInitialDelay(delay time.Duration) parameter {
	return func(p parameters) parameters {
		p.initialDelay = delay
		return p
	}
}

func WithMaxDelay(delay time.Duration) parameter {
	return func(p parameters) parameters {
		p.maxDelay = delay
		return p
	}
}

func WithExitError(errs []error) parameter {
	return func(p parameters) parameters {
		p.exitErrs = errs
		return p
	}
}

func Run[T any](ctx context.Context, fn func(context.Context) (T, error), params ...parameter) (T, error) {
	conf := parameters{
		maxTries:     defaultMaxRetries,
		initialDelay: defaultInitialDelay,
		maxDelay:     defaultMaxDelay,
	}

	for _, param := range params {
		conf = param(conf)
	}

	attempts := 0
	for {
		res, err := fn(ctx)
		if err == nil || include(conf.exitErrs, err) {
			return res, err
		}

		attempts++
		if attempts == conf.maxTries {
			return res, fmt.Errorf("max attempt exceeded: %w", err)
		}

		ticker := time.NewTimer(nextBackOff(attempts, conf.initialDelay, conf.maxDelay))
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
