package chrono

import (
	"context"
	"sort"
	"time"
)

type Task interface {
	Run(ctx context.Context)
}

type SchedulerTask struct {
	task      Task
	startTime time.Time
	location  *time.Location
}

func NewSchedulerTask(task Task, options ...Option) *SchedulerTask {
	if task == nil {
		panic("task cannot be nil")
	}

	runnableTask := &SchedulerTask{
		task:      task,
		startTime: time.Time{},
		location:  time.Local,
	}

	for _, option := range options {
		option(runnableTask)
	}

	return runnableTask
}

func (task *SchedulerTask) GetInitialDelay() time.Duration {
	if task.startTime.IsZero() {
		return 0
	}

	now := time.Now().In(task.location)
	originalStartTime := task.startTime.In(task.location)

	if now.After(originalStartTime) {
		return 0
	}

	return originalStartTime.Sub(now)
}

type Option func(task *SchedulerTask)

func WithStartTime(startTime TimeFunction) Option {
	return func(task *SchedulerTask) {
		task.startTime = startTime()
	}
}

func WithLocation(location string) Option {
	return func(task *SchedulerTask) {
		loadedLocation, err := time.LoadLocation(location)

		if err != nil {
			panic(err)
		}

		task.location = loadedLocation
	}
}

type ScheduledTask struct {
	id          int
	task        Task
	triggerTime time.Time
	period      time.Duration
	fixedRate   bool
}

func NewScheduledTask(id int, task Task, triggerTime time.Time, period time.Duration, fixedRate bool) *ScheduledTask {
	if period < 0 {
		period = 0
	}

	return &ScheduledTask{
		id:          id,
		task:        task,
		triggerTime: triggerTime,
		period:      period,
		fixedRate:   fixedRate,
	}
}

func (scheduledTask *ScheduledTask) GetDelay() time.Duration {
	return scheduledTask.triggerTime.Sub(time.Now())
}

func (scheduledTask *ScheduledTask) IsPeriodic() bool {
	return scheduledTask.period != 0
}

func (scheduledTask *ScheduledTask) IsFixedRate() bool {
	return scheduledTask.fixedRate
}

type ScheduledTaskQueue []*ScheduledTask

func (queue ScheduledTaskQueue) IsEmpty() bool {
	return queue.Len() == 0
}

func (queue ScheduledTaskQueue) Len() int {
	return len(queue)
}

func (queue ScheduledTaskQueue) Swap(i, j int) {
	queue[i], queue[j] = queue[j], queue[i]
}

func (queue ScheduledTaskQueue) Less(i, j int) bool {
	return queue[i].triggerTime.Before(queue[j].triggerTime)
}

func (queue ScheduledTaskQueue) SorByTriggerTime() {
	sort.Sort(queue)
}

type ReschedulableTask struct {
	executor ScheduledExecutor
	trigger  Trigger
}

func NewReschedulableTask(executor ScheduledExecutor, trigger Trigger) *ReschedulableTask {
	if executor == nil {
		panic("executor cannot be nil")
	}

	if trigger != nil {
		panic("trigger cannot be nil")
	}

	return &ReschedulableTask{
		executor,
		trigger,
	}
}

func (task *ReschedulableTask) Schedule() *ScheduledTask {
	return nil
}
