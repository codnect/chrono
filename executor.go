package chrono

import (
	"context"
	"errors"
	"sync"
	"time"
)

type TaskExecutor interface {
	Schedule(task Task, delay time.Duration) (ScheduledTask, error)
	ScheduleWithFixedDelay(task Task, initialDelay time.Duration, delay time.Duration) (ScheduledTask, error)
	ScheduleAtFixedRate(task Task, initialDelay time.Duration, period time.Duration) (ScheduledTask, error)
	IsShutdown() bool
	Shutdown() chan bool
}

type SimpleTaskExecutor struct {
	nextSequence          int
	isShutdown            bool
	executorMu            sync.RWMutex
	timer                 *time.Timer
	taskWaitGroup         sync.WaitGroup
	taskQueue             ScheduledTaskQueue
	newTaskChannel        chan *ScheduledRunnableTask
	rescheduleTaskChannel chan *ScheduledRunnableTask
	taskRunner            TaskRunner
	shutdownChannel       chan chan bool
}

func NewDefaultTaskExecutor() TaskExecutor {
	return NewSimpleTaskExecutor(NewDefaultTaskRunner())
}

func NewSimpleTaskExecutor(runner TaskRunner) *SimpleTaskExecutor {
	if runner == nil {
		runner = NewDefaultTaskRunner()
	}

	executor := &SimpleTaskExecutor{
		timer:                 time.NewTimer(1 * time.Hour),
		taskQueue:             make(ScheduledTaskQueue, 0),
		newTaskChannel:        make(chan *ScheduledRunnableTask),
		rescheduleTaskChannel: make(chan *ScheduledRunnableTask),
		taskRunner:            runner,
		shutdownChannel:       make(chan chan bool),
	}

	executor.timer.Stop()

	go executor.run()

	return executor
}

func (executor *SimpleTaskExecutor) Schedule(task Task, delay time.Duration) (ScheduledTask, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}

	executor.executorMu.Lock()

	if executor.isShutdown {
		executor.executorMu.Unlock()
		return nil, errors.New("no new task won't be accepted because executor is already shut down")
	}

	executor.nextSequence++
	scheduledTask, err := CreateScheduledRunnableTask(executor.nextSequence, task, executor.calculateTriggerTime(delay), 0, false)
	executor.executorMu.Unlock()

	if err != nil {
		return nil, err
	}

	executor.addNewTask(scheduledTask)

	return scheduledTask, nil
}

func (executor *SimpleTaskExecutor) ScheduleWithFixedDelay(task Task, initialDelay time.Duration, delay time.Duration) (ScheduledTask, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}

	executor.executorMu.Lock()

	if executor.isShutdown {
		executor.executorMu.Unlock()
		return nil, errors.New("no new task won't be accepted because executor is already shut down")
	}

	executor.nextSequence++
	scheduledTask, err := CreateScheduledRunnableTask(executor.nextSequence, task, executor.calculateTriggerTime(initialDelay), delay, false)
	executor.executorMu.Unlock()

	if err != nil {
		return nil, err
	}

	executor.addNewTask(scheduledTask)

	return scheduledTask, nil
}

func (executor *SimpleTaskExecutor) ScheduleAtFixedRate(task Task, initialDelay time.Duration, period time.Duration) (ScheduledTask, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}

	executor.executorMu.Lock()

	if executor.isShutdown {
		executor.executorMu.Unlock()
		return nil, errors.New("no new task won't be accepted because executor is already shut down")
	}

	executor.nextSequence++
	scheduledTask, err := CreateScheduledRunnableTask(executor.nextSequence, task, executor.calculateTriggerTime(initialDelay), period, true)
	executor.executorMu.Unlock()

	if err != nil {
		return nil, err
	}

	executor.addNewTask(scheduledTask)

	return scheduledTask, nil
}

func (executor *SimpleTaskExecutor) IsShutdown() bool {
	executor.executorMu.Lock()
	defer executor.executorMu.Unlock()
	return executor.isShutdown
}

func (executor *SimpleTaskExecutor) Shutdown() chan bool {
	executor.executorMu.Lock()
	defer executor.executorMu.Unlock()

	if executor.isShutdown {
		panic("executor is already shut down")
	}

	executor.isShutdown = true

	stoppedChan := make(chan bool)
	executor.shutdownChannel <- stoppedChan
	return stoppedChan
}

func (executor *SimpleTaskExecutor) calculateTriggerTime(delay time.Duration) time.Time {
	if delay < 0 {
		delay = 0
	}

	return time.Now().Add(delay)
}

func (executor *SimpleTaskExecutor) addNewTask(task *ScheduledRunnableTask) {
	executor.newTaskChannel <- task
}

func (executor *SimpleTaskExecutor) run() {

	for {
		executor.taskQueue.SorByTriggerTime()

		if len(executor.taskQueue) == 0 {
			executor.timer.Stop()
		} else {
			executor.timer.Reset(executor.taskQueue[0].getDelay())
		}

		for {
			select {
			case clock := <-executor.timer.C:
				executor.timer.Stop()

				taskIndex := -1
				for index, scheduledTask := range executor.taskQueue {

					if scheduledTask.triggerTime.After(clock) || scheduledTask.triggerTime.IsZero() {
						break
					}

					taskIndex = index

					if scheduledTask.IsCancelled() {
						continue
					}

					if scheduledTask.isPeriodic() && scheduledTask.isFixedRate() {
						scheduledTask.triggerTime = scheduledTask.triggerTime.Add(scheduledTask.period)
					}

					executor.startTask(scheduledTask)
				}

				executor.taskQueue = executor.taskQueue[taskIndex+1:]
			case newScheduledTask := <-executor.newTaskChannel:
				executor.timer.Stop()
				executor.taskQueue = append(executor.taskQueue, newScheduledTask)
			case rescheduledTask := <-executor.rescheduleTaskChannel:
				executor.timer.Stop()
				executor.taskQueue = append(executor.taskQueue, rescheduledTask)
			case stoppedChan := <-executor.shutdownChannel:
				executor.timer.Stop()
				executor.taskWaitGroup.Wait()
				stoppedChan <- true
				return
			}

			break
		}

	}

}

func (executor *SimpleTaskExecutor) startTask(scheduledRunnableTask *ScheduledRunnableTask) {
	executor.taskWaitGroup.Add(1)

	executor.taskRunner.Run(func(ctx context.Context) {
		defer func() {
			if executor.IsShutdown() {
				scheduledRunnableTask.Cancel()
				executor.taskWaitGroup.Done()
				return
			}

			executor.taskWaitGroup.Done()

			if !scheduledRunnableTask.isPeriodic() {
				scheduledRunnableTask.Cancel()
			} else {
				if !scheduledRunnableTask.isFixedRate() {
					scheduledRunnableTask.triggerTime = executor.calculateTriggerTime(scheduledRunnableTask.period)
					executor.rescheduleTaskChannel <- scheduledRunnableTask
				}
			}
		}()

		if scheduledRunnableTask.isPeriodic() && scheduledRunnableTask.isFixedRate() {
			executor.rescheduleTaskChannel <- scheduledRunnableTask
		}

		scheduledRunnableTask.task(ctx)
	})

}
