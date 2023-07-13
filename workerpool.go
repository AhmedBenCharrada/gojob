package gojob

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type worker struct {
	id   string
	ctx  context.Context
	stop context.CancelFunc
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
	workers     []worker

	started bool
	mutex   *sync.Mutex
}

// NewPool creates a new worker pool.
func NewPool(ctx context.Context, maxWorkers int, buffer int) *Pool {
	poolCtx, stop := context.WithCancel(ctx)

	return &Pool{
		ctx:         poolCtx,
		stop:        stop,
		jobChannel:  make(chan Job, buffer),
		maxWorkers:  orElse(maxWorkers > 0, maxWorkers, 1),
		workerCount: 0,
		workers:     make([]worker, 0, maxWorkers),
		mutex:       &sync.Mutex{},
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
	if len(p.workers)-count < 1 {
		return fmt.Errorf("can't have a worker pool without any worker")
	}

	for i := 0; i < count; i++ {
		p.mutex.Lock()
		w := p.workers[len(p.workers)-1]
		w.stop()
		p.workers = p.workers[:len(p.workers)-1]
		p.mutex.Unlock()
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

		w := worker{
			id:   uuid.NewString(),
			ctx:  ctx,
			stop: stop,
		}
		p.mutex.Lock()
		p.workers = append(p.workers, w)
		p.mutex.Unlock()
		go w.start(ctx, p.jobChannel)
	}
}

func (w *worker) start(ctx context.Context, ch <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-ch:
			job(ctx)
		}
	}
}
