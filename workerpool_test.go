package gojob_test

import (
	"context"
	"gojob"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pool := gojob.NewPool(ctx, 3, 1000)
	// add 2 workers
	err := pool.AddWorkers(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, pool.WorkerCount())

	// start the pool
	err = pool.Start()
	assert.NoError(t, err)
	defer pool.Stop()

	// re-trigger start
	pool.Start()

	// check if the number of workers did not increase due to re-triggering the start pool function.
	assert.Equal(t, 2, pool.WorkerCount())

	// add worker while the pool is on
	err = pool.AddWorkers(1)
	assert.NoError(t, err)

	// check if the worker count is increased
	assert.Equal(t, 3, pool.WorkerCount())

	// add another worker; should error with maximum allowed workers exceeded.
	err = pool.AddWorkers(1)
	assert.Error(t, err)
	assert.Equal(t, 3, pool.WorkerCount())

	// delete a worker
	err = pool.DeleteWorkers(1)
	assert.NoError(t, err)

	// check if the number of workers is decreased.
	assert.Equal(t, 2, pool.WorkerCount())

	// delete all workers; should error with can have the pool with no workers.
	err = pool.DeleteWorkers(2)
	assert.Error(t, err)

	// assert none of the workers were deleted
	assert.Equal(t, 2, pool.WorkerCount())
}

func TestWorkerPool_Push(t *testing.T) {
	ch := make(chan bool)
	j := func(ctx context.Context) {
		time.Sleep(time.Duration(rand.Intn(3)) * time.Millisecond)
		ch <- true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	pool := gojob.NewPool(ctx, 3, 2)
	time.AfterFunc(15*time.Millisecond, pool.Stop)

	err := pool.Push(j)
	assert.NoError(t, err)

	err = pool.Start()
	assert.NoError(t, err)

	// this job should be queued
	j2 := func(ctx context.Context) {
		time.Sleep(time.Duration(rand.Intn(3)) * time.Millisecond)
	}

	err = pool.Push(j2)
	assert.NoError(t, err)

	// queue is full
	err = pool.Push(func(ctx context.Context) {
		time.Sleep(time.Duration(rand.Intn(3)) * time.Millisecond)
	})
	assert.Error(t, err)

	// wait till the pool is close
	<-ctx.Done()
}
