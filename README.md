[![build](https://github.com/AhmedBenCharrada/gojob/actions/workflows/build.yml/badge.svg)](https://github.com/AhmedBenCharrada/gojob/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/AhmedBenCharrada/gojob/branch/main/graph/badge.svg?token=18JQESVC3P)](https://codecov.io/gh/AhmedBenCharrada/gojob)

---

# GOJOB

A small golang library that offers:

- A customizable retry function.

```go
fn := func(_ context.Context) (string, error) {
	return "", fmt.Errorf("breaking error")
}

res, err := gojob.Run(
    context.TODO(),
    fn,
    gojob.WithMaxTries(3),
    gojob.WithMaxDelay(2*time.Second),
    gojob.WithExitFn(func(err error) bool {
	    return err.Error() == "breaking error"
    }),
)
```

- A custom typed sync.Map wrapper.

```go
m := gojob.Map[string, string]{}
// add item
m.Put("key", "value")
// get an item
v, ok := m.Get("k")

// get size
size := m.Len()

// range through item and handle.
m.Range(func(k string, v string) error{
    // handle
})

// delete an item
m.Remove("key")
```

- A custom worker pool.

```go
// Create a worker-pool with a max of 30 worker and a job buffer size of 1000.
pool := gojob.NewPool(context.TODO(), 30, 1000)

// Add 10 workers.
err := pool.AddWorkers(10)

// Start the worker-pool.
pool.Start()

// Create a job.
job := func(ctx context.Context) {
	// do something
}

// Push the job to the worker pool.
pErr := pool.Push(job)

// Delete 5 workers.
dErr = pool.DeleteWorkers(5)

// Stop the worker-pool.
pool.Stop()

```
