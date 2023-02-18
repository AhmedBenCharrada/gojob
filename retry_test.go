package gojob_test

import (
	"context"
	"fmt"
	"gojob"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("successfully", func(t *testing.T) {
		fn := func(_ context.Context) (string, error) {
			return "done", nil
		}

		res, err := gojob.Run(context.TODO(), fn)
		assert.NoError(t, err)
		assert.Equal(t, "done", res)
	})

	t.Run("fail with an exit error", func(t *testing.T) {
		fn := func(_ context.Context) (string, error) {
			return "", fmt.Errorf("breaking error")
		}

		res, err := gojob.Run(context.TODO(), fn, gojob.WithExitFn(func(err error) bool {
			return err.Error() == "breaking error"
		}))
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("breaking error"), err, err)
		assert.Equal(t, "", res)
	})

	t.Run("fail after max attempts", func(t *testing.T) {
		tries := 0
		fn := func(_ context.Context) (string, error) {
			tries++
			return "", fmt.Errorf("error")
		}

		res, err := gojob.Run(context.TODO(), fn, gojob.WithMaxTries(3), gojob.WithMaxDelay(2*time.Second), gojob.WithExitFn(func(err error) bool {
			return err.Error() == "breaking error"
		}))
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "max attempt exceeded"))

		assert.Equal(t, "", res)

		assert.Equal(t, 3, tries)
	})

	t.Run("fail after context timeout", func(t *testing.T) {
		fn := func(_ context.Context) (string, error) {
			return "", fmt.Errorf("error")
		}

		ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
		defer cancel()

		res, err := gojob.Run(ctx, fn, gojob.WithMaxTries(10), gojob.WithInitialDelay(time.Second))
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)

		assert.Equal(t, "", res)
	})
}
