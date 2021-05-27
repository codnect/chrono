package chrono

import (
	"context"
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

func NewScheduledRunnableTask(id int, task Task, triggerTime time.Time, period time.Duration, fixedRate bool) *ScheduledRunnableTask {
	if period < 0 {
		period = 0
	}

	return &ScheduledRunnableTask{
		id:          id,
		task:        task,
		triggerTime: triggerTime,
		period:      period,
		fixedRate:   fixedRate,
	}
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

type TriggerTask struct {
	task                 Task
	currentScheduledTask *ScheduledRunnableTask
	executor             ScheduledExecutor
	triggerContext       *SimpleTriggerContext
	triggerContextMu     sync.RWMutex
	trigger              Trigger
	nextTriggerTime      time.Time
}

func NewTriggerTask(task Task, executor ScheduledExecutor, trigger Trigger) *TriggerTask {
	if executor == nil {
		panic("executor cannot be nil")
	}

	if trigger == nil {
		panic("trigger cannot be nil")
	}

	return &TriggerTask{
		task:           task,
		executor:       executor,
		triggerContext: NewSimpleTriggerContext(),
		trigger:        trigger,
	}
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

func (task *TriggerTask) Schedule() ScheduledTask {
	task.triggerContextMu.Lock()
	defer task.triggerContextMu.Unlock()

	task.nextTriggerTime = task.trigger.NextExecutionTime(task.triggerContext)

	if task.nextTriggerTime.IsZero() {
		return nil
	}

	initialDelay := task.nextTriggerTime.Sub(time.Now())

	task.currentScheduledTask = task.executor.Schedule(task.Run, initialDelay).(*ScheduledRunnableTask)

	return task
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
