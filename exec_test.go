package gojob_test

import (
	"fmt"
	"gojob"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ch := make(chan []byte, 1000)
		go printRes(ch)

		cmd := gojob.NewCommand("sh", "./testdata/test.sh")

		cmd.WithBeforeActions(func() error {
			ch <- []byte("starting")
			return nil
		})

		cmd.WithOnExecActions(func(b []byte) error {
			write(ch, b)
			return nil
		})

		cmd.WithAfterActions(func(error) error {
			close(ch)
			return nil
		})

		if err := cmd.Run(); err != nil {
			t.Fail()
		}
	})

	t.Run("with error withing before actions", func(t *testing.T) {
		cmd := gojob.NewCommand("sh", "./testdata/test.sh")

		expectedErr := fmt.Errorf("error")
		cmd.WithBeforeActions(func() error {
			return expectedErr
		})

		err := cmd.Run()
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("with error while starting the command execution", func(t *testing.T) {
		cmd := gojob.NewCommand("")
		err := cmd.Run()
		assert.Error(t, err)
	})

	t.Run("with error while executing the command", func(t *testing.T) {
		cmd := gojob.NewCommand("sh", "./testdata/test.sh")

		expectedErr := fmt.Errorf("error")
		cmd.WithOnExecActions(func(b []byte) error {
			return expectedErr
		})

		err := cmd.Run()
		assert.Error(t, err)
	})

	t.Run("with after exec error", func(t *testing.T) {
		cmd := gojob.NewCommand("sh", "./testdata/test.sh")

		expectedErr := fmt.Errorf("error")
		cmd.WithAfterActions(func(error) error {
			return expectedErr
		})

		err := cmd.Run()
		assert.ErrorIs(t, err, expectedErr)
	})
}

func write(ch chan<- []byte, b []byte) {
	if ch != nil {
		ch <- b
	}
}

func printRes(ch <-chan []byte) {
	for b := range ch {
		log.Println(string(b))
	}
}
