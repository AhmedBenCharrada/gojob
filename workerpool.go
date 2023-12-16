package gojob

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// worker status.
type status string

// worker statuses
const (
	free status = "free"
	busy status = "busy"
)

type worker struct {
	id     string
	status status
	ctx    context.Context
	stop   context.CancelFunc
}

// Job the job function signature
type Job func(context.Context)

// Pool the worker pool.
type Pool struct {
	ctx  context.Context
	stop context.CancelFunc

	jobChannel chan Job

	maxWorkers  int
	workerCount int
	workers     *Map[string, *worker]

	started bool
}

// NewPool creates a new worker pool.
func NewPool(ctx context.Context, maxWorkers int, buffer int) *Pool {
	poolCtx, stop := context.WithCancel(ctx)

	maxWorkers = orElse(maxWorkers > 0, maxWorkers, 1)
	return &Pool{
		ctx:         poolCtx,
		stop:        stop,
		jobChannel:  make(chan Job, buffer),
		maxWorkers:  maxWorkers,
		workerCount: 0,
		workers:     new(Map[string, *worker]),
	}
}

// Start triggers all workers to start working.
func (p *Pool) Start() error {
	if p.started {
		return nil
	}

	p.started = true
	workers := orElse(p.workerCount > 0, p.workerCount, 1)

	p.startWorkers(workers)
	return nil
}

// Stop stops the worker pool.
func (p *Pool) Stop() {
	p.stop()
}

// Push pushes a new job to the worker pool.
func (p *Pool) Push(job Job) error {
	select {
	case p.jobChannel <- job:
		return nil
	default:
		return fmt.Errorf("work load exceed the limit")
	}
}

// AddWorkers adds new workers to the pool.
func (p *Pool) AddWorkers(count int) error {
	if p.workerCount+count > p.maxWorkers {
		return fmt.Errorf("exceeded max allowed workers, remaining: %v", p.maxWorkers-p.workerCount)
	}

	p.workerCount += count
	if !p.started {
		return nil
	}
	p.startWorkers(count)
	return nil
}

// DeleteWorkers deletes workers from the pool.
func (p *Pool) DeleteWorkers(count int) error {
	if p.workers.Len()-count < 1 {
		return fmt.Errorf("can't have a worker pool without any worker")
	}

	var deleted int
	for deleted < count {
		p.workers.Range(func(k string, w *worker) error {
			if w.status == free {
				w.stop()
				deleted += 1
				p.workers.Remove(k)
			}

			return nil
		})
	}

	p.workerCount -= count
	return nil
}

// WorkerCount retrieves the workers count.
func (p *Pool) WorkerCount() int {
	return p.workerCount
}

func (p *Pool) startWorkers(count int) {
	for i := 0; i < count; i++ {
		ctx, stop := context.WithCancel(p.ctx)

		w := &worker{
			id:     uuid.NewString(),
			status: free,
			ctx:    ctx,
			stop:   stop,
		}

		p.workers.Put(w.id, w)
		go w.start(ctx, p.jobChannel)
	}
}

func (w *worker) start(ctx context.Context, ch <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-ch:
			w.status = busy
			job(ctx)
		}
	}
}
