package gojob

import (
	"context"
	"fmt"
	"sync"
)

// Pool the worker pool.
type Pool struct {
	ctx  context.Context
	stop context.CancelFunc

	// internal job channel
	jobChannel chan *Job

	jobStatus    *Map[*Job]
	WorkerStatus *Map[*Worker]
}

// NewPool creates a new worker pool.
func NewPool(ctx context.Context, buffer int) *Pool {
	c, stop := context.WithCancel(ctx)
	return &Pool{
		ctx:        c,
		stop:       stop,
		jobChannel: make(chan *Job, buffer),
		jobStatus: &Map[*Job]{
			m:     make(map[string]*Job),
			mutex: sync.RWMutex{},
		},
		WorkerStatus: &Map[*Worker]{
			m:     make(map[string]*Worker),
			mutex: sync.RWMutex{},
		},
	}
}

// AddWorkers add a set of workers to the worker pool.
func (p *Pool) AddWorkers(workers []*Worker) {
	for i := range workers {
		w := workers[i]
		_ = p.WorkerStatus.Push(w)
		go w.Start(p.ctx, p.jobChannel)
	}
}

// PushJob pushes a job to the pool.
func (p *Pool) PushJob(job *Job) error {
	if err := p.jobStatus.Push(job); err != nil {
		return err
	}

	p.jobChannel <- job
	return nil
}

// GetJobInfo extracts the job info.
func (p *Pool) GetJobInfo(id string) *Job {
	job, ok := p.jobStatus.Get(id)
	if !ok {
		return nil
	}

	return job
}

// CancelJob cancels a job that is still in the queue.
func (p *Pool) CancelJob(id string) error {
	job, ok := p.jobStatus.Get(id)
	if !ok {
		return fmt.Errorf("job not found")
	}

	if job.Status == Done {
		return nil
	}

	if job.stop != nil {
		job.stop()
	}

	return nil
}
