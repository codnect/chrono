package chrono

import (
	"time"
)

type TaskScheduler interface {
	Schedule(task Task, options ...Option) (ScheduledTask, error)
	ScheduleWithCron(task Task, expression string, options ...Option) (ScheduledTask, error)
	ScheduleWithFixedDelay(task Task, delay time.Duration, options ...Option) (ScheduledTask, error)
	ScheduleAtFixedRate(task Task, period time.Duration, options ...Option) (ScheduledTask, error)
	IsShutdown() bool
	Shutdown() chan bool
}

type SimpleTaskScheduler struct {
	taskExecutor TaskExecutor
}

func NewSimpleTaskScheduler(executor TaskExecutor) *SimpleTaskScheduler {

	if executor == nil {
		executor = NewDefaultTaskExecutor()
	}

	scheduler := &SimpleTaskScheduler{
		taskExecutor: executor,
	}

	return scheduler
}

func NewDefaultTaskScheduler() TaskScheduler {
	return NewSimpleTaskScheduler(NewDefaultTaskExecutor())
}

func (scheduler *SimpleTaskScheduler) Schedule(task Task, options ...Option) (ScheduledTask, error) {
	schedulerTask, err := CreateSchedulerTask(task, options...)

	if err != nil {
		return nil, err
	}

	return scheduler.taskExecutor.Schedule(task, schedulerTask.GetInitialDelay())
}

func (scheduler *SimpleTaskScheduler) ScheduleWithCron(task Task, expression string, options ...Option) (ScheduledTask, error) {
	var schedulerTask *SchedulerTask
	var err error

	schedulerTask, err = CreateSchedulerTask(task, options...)

	if err != nil {
		return nil, err
	}

	var cronTrigger *CronTrigger
	cronTrigger, err = CreateCronTrigger(expression, schedulerTask.location)

	if err != nil {
		return nil, err
	}

	var triggerTask *TriggerTask
	triggerTask, err = CreateTriggerTask(schedulerTask.task, scheduler.taskExecutor, cronTrigger)

	if err != nil {
		return nil, err
	}

	return triggerTask.Schedule()
}

func (scheduler *SimpleTaskScheduler) ScheduleWithFixedDelay(task Task, delay time.Duration, options ...Option) (ScheduledTask, error) {
	schedulerTask, err := CreateSchedulerTask(task, options...)

	if err != nil {
		return nil, err
	}

	return scheduler.taskExecutor.ScheduleWithFixedDelay(schedulerTask.task, schedulerTask.GetInitialDelay(), delay)
}

func (scheduler *SimpleTaskScheduler) ScheduleAtFixedRate(task Task, period time.Duration, options ...Option) (ScheduledTask, error) {
	schedulerTask, err := CreateSchedulerTask(task, options...)

	if err != nil {
		return nil, err
	}

	return scheduler.taskExecutor.ScheduleAtFixedRate(schedulerTask.task, schedulerTask.GetInitialDelay(), period)
}

func (scheduler *SimpleTaskScheduler) IsShutdown() bool {
	return scheduler.taskExecutor.IsShutdown()
}

func (scheduler *SimpleTaskScheduler) Shutdown() chan bool {
	return scheduler.taskExecutor.Shutdown()
}
