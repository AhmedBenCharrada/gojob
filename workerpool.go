package gojob

import (
	"context"
	"fmt"
	"sync"
)

type worker struct {
	id   int
	ctx  context.Context
	stop context.CancelFunc
}

type Job func(context.Context)

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

func New(ctx context.Context, maxWorkers int, buffer int) *Pool {
	poolCtx, stop := context.WithCancel(ctx)
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	return &Pool{
		ctx:         poolCtx,
		stop:        stop,
		jobChannel:  make(chan Job, buffer),
		maxWorkers:  maxWorkers,
		workerCount: 0,
		workers:     make([]worker, maxWorkers),
		mutex:       &sync.Mutex{},
	}
}

func (p *Pool) Start() error {
	if p.started {
		return nil
	}

	p.started = true
	if p.workerCount < 1 {
		return p.AddWorkers(1)
	}

	return p.AddWorkers(p.workerCount)
}

func (p *Pool) Stop() {
	p.stop()
}

func (p *Pool) Push(job Job) error {
	select {
	case p.jobChannel <- job:
		return nil
	default:
		return fmt.Errorf("work load exceed the limit")
	}
}

func (p *Pool) AddWorkers(count int) error {
	if p.workerCount+count > p.maxWorkers {
		return fmt.Errorf("exceeded max allowed workers, remaining: %v", p.maxWorkers-p.workerCount)
	}

	p.workerCount += count
	if !p.started {
		return nil
	}

	for i := 0; i < count; i++ {
		ctx, stop := context.WithCancel(p.ctx)

		w := worker{
			id:   p.workerCount,
			ctx:  ctx,
			stop: stop,
		}
		p.mutex.Lock()
		p.workers = append(p.workers, w)
		p.mutex.Unlock()
		go w.start(ctx, p.jobChannel)
	}

	return nil
}

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
