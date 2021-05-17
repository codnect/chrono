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
		executor = NewScheduledTaskExecutor()
	}

	scheduler := &SimpleScheduler{
		executor: executor,
	}

	return scheduler
}

func (scheduler *SimpleScheduler) Schedule(task Task, options ...Option) *ScheduledTask {
	//schedulerTask := NewSchedulerTask(task, options...)
	return nil
}

func (scheduler *SimpleScheduler) ScheduleWithCron(task Task, expression string, options ...Option) *ScheduledTask {
	//	schedulerTask := NewSchedulerTask(task, options...)
	return nil
}

func (scheduler *SimpleScheduler) ScheduleWithFixedDelay(task Task, delay time.Duration, options ...Option) *ScheduledTask {
	//schedulerTask := NewSchedulerTask(task, options...)
	return nil
}

func (scheduler *SimpleScheduler) ScheduleAtFixedRate(task Task, period time.Duration, options ...Option) *ScheduledTask {
	schedulerTask := NewSchedulerTask(task, options...)

	//	trigger := NewPeriodicTrigger(period, 0, true)
	//reschedulableTask := NewReschedulableTask(scheduler.executor, trigger)

	return scheduler.executor.ScheduleAtWithRate(schedulerTask.task, 0, period)
}

func (scheduler *SimpleScheduler) Terminate() {

}
