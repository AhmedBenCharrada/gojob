package main

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 21*time.Second)
	defer cancel()

	pool := New(ctx, 5, 100)

	err := pool.Start()
	assert.NoError(t, err)

	err = pool.AddWorkers(4)
	assert.NoError(t, err)

	jobCount := 10
	ch := make(chan bool, jobCount)
	wg := &sync.WaitGroup{}
	wg.Add(jobCount)

	jobs := getTestJobs(wg, ch, jobCount, 1000)

	go func() {
		wg.Wait()
		close(ch)
	}()

	for _, job := range jobs {
		err := pool.Push(job)
		assert.NoError(t, err)
	}

loop:
	for {
		select {
		case <-ctx.Done():
			t.Fail()
			break loop
		case r := <-ch:
			if r {
				jobCount--
				continue
			}

			break loop
		}
	}

	assert.Equal(t, 0, jobCount)
}

func getTestJobs(wg *sync.WaitGroup, ch chan bool, count int, delayInMs int) []Job {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	jobs := make([]Job, 0, count)

	for i := 0; i < count; i++ {
		randDelay := delayInMs + rand.Intn(1000)
		jobs = append(jobs, func(ctx context.Context) {
			defer wg.Done()
			time.Sleep(time.Duration(randDelay) * time.Millisecond)
			ch <- true
		})
	}

	return jobs
}
