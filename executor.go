package chrono

import (
	"sync"
	"time"
)

type ScheduledExecutor interface {
	Schedule(task Task, delay time.Duration) *ScheduledTask
	ScheduleWithFixedDelay(task Task, initialDelay time.Duration, delay time.Duration) *ScheduledTask
	ScheduleAtWithRate(task Task, initialDelay time.Duration, period time.Duration) *ScheduledTask
}

type ScheduledTaskExecutor struct {
	timer          *time.Timer
	taskQueue      ScheduledTaskQueue
	taskQueueMu    sync.RWMutex
	newTaskChannel chan *ScheduledTask
}

func NewScheduledTaskExecutor() *ScheduledTaskExecutor {
	executor := &ScheduledTaskExecutor{
		timer:          time.NewTimer(1 * time.Hour),
		taskQueue:      make(ScheduledTaskQueue, 0),
		taskQueueMu:    sync.RWMutex{},
		newTaskChannel: make(chan *ScheduledTask),
	}

	executor.timer.Stop()

	go executor.run()

	return executor
}

func (executor *ScheduledTaskExecutor) Schedule(task Task, delay time.Duration) *ScheduledTask {
	if task == nil {
		panic("task cannot be nil")
	}

	scheduledTask := NewScheduledTask(task, executor.calculateTriggerTime(delay), 0)
	executor.addNewTask(scheduledTask)

	return scheduledTask
}

func (executor *ScheduledTaskExecutor) ScheduleWithFixedDelay(task Task, initialDelay time.Duration, delay time.Duration) *ScheduledTask {
	if task == nil {
		panic("task cannot be nil")
	}

	scheduledTask := NewScheduledTask(task, executor.calculateTriggerTime(initialDelay), delay)
	executor.addNewTask(scheduledTask)

	return scheduledTask
}

func (executor *ScheduledTaskExecutor) ScheduleAtWithRate(task Task, initialDelay time.Duration, period time.Duration) *ScheduledTask {
	if task == nil {
		panic("task cannot be nil")
	}

	scheduledTask := NewScheduledTask(task, executor.calculateTriggerTime(initialDelay), period)
	executor.addNewTask(scheduledTask)

	return scheduledTask
}

func (executor *ScheduledTaskExecutor) calculateTriggerTime(delay time.Duration) time.Time {
	if delay < 0 {
		delay = 0
	}

	return time.Now().Add(delay)
}

func (executor *ScheduledTaskExecutor) addNewTask(task *ScheduledTask) {
	executor.newTaskChannel <- task
}

func (executor *ScheduledTaskExecutor) run() {

	for {

		if executor.taskQueue.IsEmpty() {
			executor.timer.Stop()
		} else {
			executor.timer.Reset(executor.taskQueue[0].GetDelay())
		}

		for {
			select {
			case clock := <-executor.timer.C:
				executor.taskQueueMu.Lock()

				var index int
				var task *ScheduledTask
				for index, task = range executor.taskQueue {
					if task.triggerTime.After(clock) || task.triggerTime.IsZero() {
						break
					}
				}

				executor.taskQueue = executor.taskQueue[index:]
				executor.taskQueueMu.Unlock()
			case newScheduledTask := <-executor.newTaskChannel:
				executor.timer.Stop()

				executor.taskQueueMu.Lock()
				executor.taskQueue = append(executor.taskQueue, newScheduledTask)
				executor.taskQueue.SorByTriggerTime()
				executor.taskQueueMu.Unlock()
			}

			break
		}

	}

}
