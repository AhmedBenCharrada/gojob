package gojob

import "os/exec"

// Command external command to run.
type Command struct {
	cmd        string
	args       []string
	before     []beforeAction
	after      []afterAction
	execAction []writeAction
}

type beforeAction func() error
type afterAction func(error) error
type writeAction func([]byte) error

// NewCommand creates a new command.
func NewCommand(cmd string, args ...string) *Command {
	return &Command{cmd: cmd, args: args}
}

func (c *Command) WithBeforeActions(actions ...beforeAction) {
	c.before = actions
}

func (c *Command) WithAfterActions(actions ...afterAction) {
	c.after = actions
}

func (c *Command) WithOnExecActions(actions ...writeAction) {
	c.execAction = actions
}

func (c *Command) Run() error {
	out := newWriter(c.execAction...)

	cmd := exec.Command(c.cmd, c.args...)
	cmd.Stdout = out

	for _, bc := range c.before {
		if err := bc(); err != nil {
			return err
		}
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	err := cmd.Wait()
	for _, ac := range c.after {
		if err := ac(err); err != nil {
			return err
		}
	}

	return err
}

type writer struct {
	actions []writeAction
}

func newWriter(actions ...writeAction) *writer {
	return &writer{actions: actions}
}

func (w *writer) Write(in []byte) (int, error) {
	for _, action := range w.actions {
		if err := action(in); err != nil {
			return -1, err
		}
	}

	return len(in), nil
}
