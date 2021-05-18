package chrono

import (
	"time"
)

type Scheduler interface {
	Schedule(task Task, options ...Option) *ScheduledTask
	ScheduleWithCron(task Task, expression string, options ...Option) *ScheduledTask
	ScheduleWithFixedDelay(task Task, delay time.Duration, options ...Option) *ScheduledTask
	ScheduleAtFixedRate(task Task, period time.Duration, options ...Option) *ScheduledTask
	Terminate()
}

type SimpleScheduler struct {
	executor ScheduledExecutor
}

func NewScheduler(executor ScheduledExecutor) Scheduler {

	if executor == nil {
		executor = NewDefaultScheduledExecutor()
	}

	scheduler := &SimpleScheduler{
		executor: executor,
	}

	return scheduler
}

func (scheduler *SimpleScheduler) Schedule(task Task, options ...Option) *ScheduledTask {
	schedulerTask := NewSchedulerTask(task, options...)
	return scheduler.executor.Schedule(task, schedulerTask.GetInitialDelay())
}

func (scheduler *SimpleScheduler) ScheduleWithCron(task Task, expression string, options ...Option) *ScheduledTask {
	schedulerTask := NewSchedulerTask(task, options...)
	return scheduler.executor.Schedule(task, schedulerTask.GetInitialDelay())
}

func (scheduler *SimpleScheduler) ScheduleWithFixedDelay(task Task, delay time.Duration, options ...Option) *ScheduledTask {
	schedulerTask := NewSchedulerTask(task, options...)
	return scheduler.executor.ScheduleWithFixedDelay(schedulerTask.task, schedulerTask.GetInitialDelay(), delay)
}

func (scheduler *SimpleScheduler) ScheduleAtFixedRate(task Task, period time.Duration, options ...Option) *ScheduledTask {
	schedulerTask := NewSchedulerTask(task, options...)
	return scheduler.executor.ScheduleAtWithRate(schedulerTask.task, schedulerTask.GetInitialDelay(), period)
}

func (scheduler *SimpleScheduler) Terminate() {

}
