package gojob

import (
	"context"
	"time"
)

// JobType ..
type JobType string

// JobFn the job function signature
type JobFn func(context.Context)

// JobStatus the job execution status
type JobStatus string

// Job execution statuses
const (
	Pending JobStatus = "pending"
	Running JobStatus = "running"
	Done    JobStatus = "done"
)

// JobHistory ..
type JobHistory struct {
	Status    JobStatus
	WorkerID  string
	Timestamp uint64
}

// Job ..
type Job struct {
	id   string
	kind JobType
	meta map[string]string
	fn   JobFn

	Status    JobStatus
	WorkerID  string
	Timestamp uint64
	History   []JobHistory

	ctx  context.Context
	stop context.CancelFunc
}

// NewJob ..
func NewJob(id string, jobType JobType, meta map[string]string, fn JobFn) *Job {
	now := uint64(time.Now().Unix())
	return &Job{
		id:        id,
		kind:      jobType,
		meta:      meta,
		fn:        fn,
		Status:    Pending,
		Timestamp: now,
		History: []JobHistory{
			{
				Status:    Pending,
				Timestamp: now,
			},
		},
	}
}

// ID ..
func (j *Job) ID() string {
	return j.id
}

// Type the job type.
func (j *Job) Type() JobType {
	return j.kind
}

// Meta ..
func (j *Job) Meta() map[string]string {
	return j.meta
}

// Run ..
func (j *Job) Run(ctx context.Context) {
	j.ctx, j.stop = context.WithCancel(ctx)
	j.fn(ctx)
}
