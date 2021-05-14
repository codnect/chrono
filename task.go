package scheduler

import (
	"context"
	"time"
)

type Task func(ctx context.Context)

type ScheduledTask struct {
	task              Task
	isAsync           bool
	initialDelay      time.Duration
	nextExecutionTime time.Time
	fixedRate         time.Duration
	fixedDelay        time.Duration
	location          *time.Location
}

func NewScheduledTask(task Task, options ...Option) *ScheduledTask {
	if task == nil {
		panic("task cannot be nil")
	}

	scheduledTask := &ScheduledTask{
		task:         task,
		initialDelay: 0,
		fixedRate:    -1,
		fixedDelay:   -1,
		location:     time.Local,
	}

	for _, option := range options {
		option(scheduledTask)
	}

	return scheduledTask
}

func (task *ScheduledTask) Execute(ctx context.Context) {

}
