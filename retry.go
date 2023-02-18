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
	exitFn                 func(error) bool
}

type parameter func(*parameters)

func WithMaxTries(tries int) parameter {
	return func(p *parameters) {
		p.maxTries = tries
	}
}

func WithInitialDelay(delay time.Duration) parameter {
	return func(p *parameters) {
		p.initialDelay = delay
	}
}

func WithMaxDelay(delay time.Duration) parameter {
	return func(p *parameters) {
		p.maxDelay = delay
	}
}

func WithExitFn(fn func(error) bool) parameter {
	return func(p *parameters) {
		p.exitFn = fn
	}
}

type response[T any] struct {
	data T
	err  error
}

func Run[T any](ctx context.Context, fn func(context.Context) (T, error), params ...parameter) (T, error) {
	conf := &parameters{
		maxTries:     defaultMaxRetries,
		initialDelay: defaultInitialDelay,
		maxDelay:     defaultMaxDelay,
	}

	for _, param := range params {
		param(conf)
	}

	ch := make(chan response[T])
	go run(ctx, conf, fn, ch)

	select {
	case res := <-ch:
		return res.data, res.err
	case <-ctx.Done():
		return *new(T), ctx.Err()
	}
}

func run[T any](ctx context.Context, conf *parameters, fn func(context.Context) (T, error), ch chan<- response[T]) {
	attempts := 0
	for {
		res, err := fn(ctx)
		if err == nil || (conf.exitFn != nil && conf.exitFn(err)) {
			ch <- response[T]{
				data: res,
				err:  err,
			}
			return
		}

		attempts++
		if attempts == conf.maxTries {
			ch <- response[T]{
				data: res,
				err:  fmt.Errorf("max attempt exceeded: %w", err),
			}
			return
		}

		ticker := time.NewTimer(nextBackOff(attempts, conf.initialDelay, conf.maxDelay))
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			ch <- response[T]{
				data: res,
				err:  ctx.Err(),
			}
			return
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
