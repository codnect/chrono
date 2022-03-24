package chrono

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Task func(ctx context.Context)

type SchedulerTask struct {
	task      Task
	startTime time.Time
	location  *time.Location
}

func CreateSchedulerTask(task Task, options ...Option) (*SchedulerTask, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}

	runnableTask := &SchedulerTask{
		task:      task,
		startTime: time.Time{},
		location:  time.Local,
	}

	for _, option := range options {
		err := option(runnableTask)

		if err != nil {
			return nil, err
		}
	}

	return runnableTask, nil
}

func (task *SchedulerTask) GetInitialDelay() time.Duration {
	if task.startTime.IsZero() {
		return 0
	}

	now := time.Now().In(task.location)
	diff := time.Date(task.startTime.Year(), task.startTime.Month(), task.startTime.Day(), task.startTime.Hour(), task.startTime.Minute(), task.startTime.Second(), 0, time.Local).Sub(
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local))

	if diff < 0 {
		return 0
	}

	return diff
}

type Option func(task *SchedulerTask) error

func WithTime(t time.Time) Option {
	return func(task *SchedulerTask) error {
		task.startTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)

		if t.Location() != nil && t.Location() != time.Local {
			task.location = t.Location()
		}

		return nil
	}
}

// Deprecated: Use WithTime instead.
func WithStartTime(year int, month time.Month, day, hour, min, sec int) Option {
	return func(task *SchedulerTask) error {
		task.startTime = time.Date(year, month, day, hour, min, sec, 0, time.Local)
		return nil
	}
}

func WithLocation(location string) Option {
	return func(task *SchedulerTask) error {
		loadedLocation, err := time.LoadLocation(location)

		if err != nil {
			return fmt.Errorf("location not loaded : %s", location)
		}

		task.location = loadedLocation
		return nil
	}
}

type ScheduledTask interface {
	Cancel()
	IsCancelled() bool
}

type ScheduledRunnableTask struct {
	id          int
	task        Task
	taskMu      sync.RWMutex
	triggerTime time.Time
	period      time.Duration
	fixedRate   bool
	cancelled   bool
}

func CreateScheduledRunnableTask(id int, task Task, triggerTime time.Time, period time.Duration, fixedRate bool) (*ScheduledRunnableTask, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}

	if period < 0 {
		period = 0
	}

	return &ScheduledRunnableTask{
		id:          id,
		task:        task,
		triggerTime: triggerTime,
		period:      period,
		fixedRate:   fixedRate,
	}, nil
}

func (scheduledRunnableTask *ScheduledRunnableTask) Cancel() {
	scheduledRunnableTask.taskMu.Lock()
	defer scheduledRunnableTask.taskMu.Unlock()
	scheduledRunnableTask.cancelled = true
}

func (scheduledRunnableTask *ScheduledRunnableTask) IsCancelled() bool {
	scheduledRunnableTask.taskMu.Lock()
	defer scheduledRunnableTask.taskMu.Unlock()
	return scheduledRunnableTask.cancelled
}

func (scheduledRunnableTask *ScheduledRunnableTask) getDelay() time.Duration {
	return scheduledRunnableTask.triggerTime.Sub(time.Now())
}

func (scheduledRunnableTask *ScheduledRunnableTask) isPeriodic() bool {
	return scheduledRunnableTask.period != 0
}

func (scheduledRunnableTask *ScheduledRunnableTask) isFixedRate() bool {
	return scheduledRunnableTask.fixedRate
}

type ScheduledTaskQueue []*ScheduledRunnableTask

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

type TriggerTask struct {
	task                 Task
	currentScheduledTask *ScheduledRunnableTask
	executor             TaskExecutor
	triggerContext       *SimpleTriggerContext
	triggerContextMu     sync.RWMutex
	trigger              Trigger
	nextTriggerTime      time.Time
}

func CreateTriggerTask(task Task, executor TaskExecutor, trigger Trigger) (*TriggerTask, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}

	if executor == nil {
		return nil, errors.New("executor cannot be nil")
	}

	if trigger == nil {
		return nil, errors.New("trigger cannot be nil")
	}

	return &TriggerTask{
		task:           task,
		executor:       executor,
		triggerContext: NewSimpleTriggerContext(),
		trigger:        trigger,
	}, nil
}

func (task *TriggerTask) Cancel() {
	task.triggerContextMu.Lock()
	defer task.triggerContextMu.Unlock()
	task.currentScheduledTask.Cancel()
}

func (task *TriggerTask) IsCancelled() bool {
	task.triggerContextMu.Lock()
	defer task.triggerContextMu.Unlock()
	return task.currentScheduledTask.IsCancelled()
}

func (task *TriggerTask) Schedule() (ScheduledTask, error) {
	task.triggerContextMu.Lock()
	defer task.triggerContextMu.Unlock()

	task.nextTriggerTime = task.trigger.NextExecutionTime(task.triggerContext)

	if task.nextTriggerTime.IsZero() {
		return nil, errors.New("could not schedule task because of the fact that schedule time is zero")
	}

	initialDelay := task.nextTriggerTime.Sub(time.Now())

	currentScheduledTask, err := task.executor.Schedule(task.Run, initialDelay)

	if err != nil {
		return nil, err
	}

	task.currentScheduledTask = currentScheduledTask.(*ScheduledRunnableTask)
	return task, nil
}

func (task *TriggerTask) Run(ctx context.Context) {
	task.triggerContextMu.Lock()

	executionTime := time.Now()
	task.task(ctx)
	completionTime := time.Now()

	task.triggerContext.Update(completionTime, executionTime, task.nextTriggerTime)
	task.triggerContextMu.Unlock()

	if !task.IsCancelled() {
		task.Schedule()
	}
}
