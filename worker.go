package gojob

import "context"

// WorkerStatus ..
type WorkerStatus string

// the workers execution statuses.
const (
	Busy WorkerStatus = "busy"
	Free WorkerStatus = "free"

	Stopped WorkerStatus = "stopped"
)

// Worker ..
type Worker struct {
	id      string
	jobType JobType
	Meta    map[string]string
	Status  WorkerStatus

	ctx  context.Context
	stop context.CancelFunc
}

// NewWorker creates a new worker.
func NewWorker(id string, jobType JobType, meta map[string]string) *Worker {
	return &Worker{
		id:      id,
		jobType: jobType,
		Meta:    meta,
		Status:  Free,
	}
}

// ID ..
func (w *Worker) ID() string {
	return w.id
}

// Start ..
func (w *Worker) Start(ctx context.Context, ch <-chan *Job) {
	w.ctx, w.stop = context.WithCancel(ctx)

	for {
		select {
		case <-w.ctx.Done():
			return
		case job := <-ch:
			w.Status = Busy
			job.Status = Running
			job.Run(w.ctx)
			job.Status = Done
			w.Status = Free
		}
	}
}

// Stop ..
func (w *Worker) Stop() {
	if w.stop != nil {
		w.stop()
	}
}
