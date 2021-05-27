package chrono

import (
	"time"
)

type Scheduler interface {
	Schedule(task Task, options ...Option) (ScheduledTask, error)
	ScheduleWithCron(task Task, expression string, options ...Option) (ScheduledTask, error)
	ScheduleWithFixedDelay(task Task, delay time.Duration, options ...Option) (ScheduledTask, error)
	ScheduleAtFixedRate(task Task, period time.Duration, options ...Option) (ScheduledTask, error)
	IsShutdown() bool
	Shutdown() chan bool
}

type SimpleScheduler struct {
	executor ScheduledExecutor
}

func NewSimpleScheduler(executor ScheduledExecutor) *SimpleScheduler {

	if executor == nil {
		executor = NewDefaultScheduledExecutor()
	}

	scheduler := &SimpleScheduler{
		executor: executor,
	}

	return scheduler
}

func NewDefaultScheduler() Scheduler {
	return NewSimpleScheduler(NewDefaultScheduledExecutor())
}

func (scheduler *SimpleScheduler) Schedule(task Task, options ...Option) (ScheduledTask, error) {
	schedulerTask := NewSchedulerTask(task, options...)
	return scheduler.executor.Schedule(task, schedulerTask.GetInitialDelay())
}

func (scheduler *SimpleScheduler) ScheduleWithCron(task Task, expression string, options ...Option) (ScheduledTask, error) {
	schedulerTask := NewSchedulerTask(task, options...)
	triggerTask := NewTriggerTask(schedulerTask.task, scheduler.executor, NewCronTrigger(expression, schedulerTask.location))
	return triggerTask.Schedule()
}

func (scheduler *SimpleScheduler) ScheduleWithFixedDelay(task Task, delay time.Duration, options ...Option) (ScheduledTask, error) {
	schedulerTask := NewSchedulerTask(task, options...)
	return scheduler.executor.ScheduleWithFixedDelay(schedulerTask.task, schedulerTask.GetInitialDelay(), delay)
}

func (scheduler *SimpleScheduler) ScheduleAtFixedRate(task Task, period time.Duration, options ...Option) (ScheduledTask, error) {
	schedulerTask := NewSchedulerTask(task, options...)
	return scheduler.executor.ScheduleAtFixedRate(schedulerTask.task, schedulerTask.GetInitialDelay(), period)
}

func (scheduler *SimpleScheduler) IsShutdown() bool {
	return scheduler.executor.IsShutdown()
}

func (scheduler *SimpleScheduler) Shutdown() chan bool {
	return scheduler.executor.Shutdown()
}
