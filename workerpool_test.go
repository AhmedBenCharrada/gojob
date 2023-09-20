package gojob_test

import (
	"context"
	"gojob"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()

	pool := gojob.NewPool(ctx, 5)

	err := pool.PushJob(gojob.NewJob("job-0", gojob.JobType("type"), map[string]string{"owner": "system"}, func(ctx context.Context) {}))
	assert.NoError(t, err)

	j0 := pool.GetJobInfo("job-0")
	assert.Equal(t, "job-0", j0.ID())
	assert.Equal(t, gojob.Pending, j0.Status)
	assert.NotEmpty(t, gojob.Pending, j0.Meta())
	assert.Equal(t, "system", j0.Meta()["owner"])

	assert.NoError(t, pool.CancelJob("job-0"))

	wr := gojob.NewWorker(uuid.NewString(), gojob.JobType("type"), nil)
	pool.AddWorkers([]*gojob.Worker{wr})

	job := gojob.NewJob("job-1", gojob.JobType("type"), nil, func(ctx context.Context) {
		for i := 0; i < 5; i++ {
			log.Println("job-1", i)
			time.Sleep(time.Second)
		}
	})

	err = pool.PushJob(job)
	assert.NoError(t, err)

	err = pool.PushJob(gojob.NewJob("", gojob.JobType("type"), nil, func(ctx context.Context) {}))
	assert.Error(t, err, "invalid job (missing job ID)")

	time.Sleep(time.Second)

	w, ok := pool.WorkerStatus.Get(wr.ID())
	assert.True(t, ok)
	assert.Equal(t, gojob.Busy, w.Status)

	j := pool.GetJobInfo("job-1")
	assert.Equal(t, "job-1", j.ID())
	assert.Equal(t, gojob.Running, j.Status)

	assert.Nil(t, pool.GetJobInfo("job-2"))
	assert.Error(t, pool.CancelJob("job-2"))

	assert.NoError(t, pool.CancelJob("job-1"))

}
