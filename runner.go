package chrono

import (
	"context"
)

type TaskRunner interface {
	Run(task Task)
}

type SimpleTaskRunner struct {
}

func NewDefaultTaskRunner() TaskRunner {
	return NewSimpleTaskRunner()
}

func NewSimpleTaskRunner() *SimpleTaskRunner {
	return &SimpleTaskRunner{}
}

func (runner *SimpleTaskRunner) Run(task Task) {
	go func() {
		task(context.Background())
	}()
}
