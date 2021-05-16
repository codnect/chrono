package chrono

import (
	"context"
	"sync"
	"time"
)

type Scheduler interface {
	Start()
	IsActive() bool
	Schedule(task Task, options ...Option) ScheduledTask
	ScheduleWithCron(task Task, expression string, options ...Option) ScheduledTask
	ScheduleWithFixedDelay(task Task, delay time.Duration, options ...Option) ScheduledTask
	ScheduleWithFixedRate(task Task, period time.Duration, options ...Option) ScheduledTask
	Terminate()
}

type SimpleScheduler struct {
	isActive          bool
	mu                sync.RWMutex
	timer             *time.Timer
	nextTaskId        int
	newTaskChannel    chan *RunnableTask
	removeTaskChannel chan int
	tasks             Tasks
	wg                *sync.WaitGroup
	cancelFunctions   []context.CancelFunc
}

func NewScheduler() Scheduler {
	return &SimpleScheduler{
		isActive:        false,
		mu:              sync.RWMutex{},
		nextTaskId:      0,
		newTaskChannel:  make(chan *RunnableTask),
		tasks:           make(Tasks, 0),
		wg:              &sync.WaitGroup{},
		cancelFunctions: make([]context.CancelFunc, 0),
	}
}

func (scheduler *SimpleScheduler) Start() {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	if scheduler.isActive {
		return
	}

	scheduler.isActive = true

	go scheduler.run()
}

func (scheduler *SimpleScheduler) IsActive() bool {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	return scheduler.isActive
}

func (scheduler *SimpleScheduler) Schedule(task Task, options ...Option) ScheduledTask {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	scheduler.nextTaskId++
	scheduledTask := NewRunnableTask(scheduler.nextTaskId, task, options...)

	return scheduledTask
	/*

		ctxWithCancel, cancel := context.WithCancel(ctx)
		scheduler.wg.Add(1)

		scheduler.cancelFunctions = append(scheduler.cancelFunctions, cancel)
		go scheduler.execute(ctxWithCancel, task, 1*time.Second)
	*/
}

func (scheduler *SimpleScheduler) ScheduleWithCron(task Task, expression string, options ...Option) ScheduledTask {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	scheduler.nextTaskId++
	scheduledTask := NewRunnableTask(scheduler.nextTaskId, task, options...)
	scheduledTask.trigger = NewCronTrigger(expression)

	return scheduledTask
}

func (scheduler *SimpleScheduler) ScheduleWithFixedDelay(task Task, delay time.Duration, options ...Option) ScheduledTask {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	scheduler.nextTaskId++
	scheduledTask := NewRunnableTask(scheduler.nextTaskId, task, options...)

	initialDelay := time.Now().In(scheduledTask.location).Sub(scheduledTask.startTime.In(scheduledTask.location))
	scheduledTask.trigger = NewPeriodicTrigger(delay, initialDelay, false)

	return scheduledTask
}

func (scheduler *SimpleScheduler) ScheduleWithFixedRate(task Task, period time.Duration, options ...Option) ScheduledTask {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	scheduler.nextTaskId++
	scheduledTask := NewRunnableTask(scheduler.nextTaskId, task, options...)

	initialDelay := time.Now().In(scheduledTask.location).Sub(scheduledTask.startTime.In(scheduledTask.location))
	scheduledTask.trigger = NewPeriodicTrigger(period, initialDelay, true)

	return scheduledTask
}

func (scheduler *SimpleScheduler) addTask(scheduledTask *RunnableTask) {
	if scheduler.isActive {
		scheduler.newTaskChannel <- scheduledTask
	} else {
		scheduler.tasks = append(scheduler.tasks, scheduledTask)
	}
}

func (scheduler *SimpleScheduler) Terminate() {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	for _, cancelFunction := range scheduler.cancelFunctions {
		cancelFunction()
	}

	scheduler.wg.Wait()
}

func (scheduler *SimpleScheduler) run() {

	now := time.Now()
	scheduler.tasks.UpdateNextExecutionTimes(now)

	for {
		scheduler.tasks.SortByNextExecutionTime()

		if scheduler.timer == nil {
			scheduler.timer = time.NewTimer(0)
		} else {

			if scheduler.tasks.IsEmpty() {
				scheduler.timer.Stop()
			} else {
				scheduler.timer.Reset(0)
			}

		}

		for {

			select {
			case now = <-scheduler.timer.C:

			case newTask := <-scheduler.newTaskChannel:
				scheduler.timer.Stop()
				//newTask.updateNextExecutionTime(time.Now())
				scheduler.tasks = append(scheduler.tasks, newTask)
			case taskId := <-scheduler.removeTaskChannel:
				scheduler.timer.Stop()
				now = time.Now()
				scheduler.removeTask(taskId)
			}

			break
		}
	}
}

func (scheduler *SimpleScheduler) removeTask(taskId int) {
	taskIndex := -1

	for index, task := range scheduler.tasks {
		if task.id != taskId {
			continue
		}

		taskIndex = index
	}

	if taskIndex == -1 {
		return
	}

	scheduler.tasks = append(scheduler.tasks[:taskIndex], scheduler.tasks[taskIndex+1:]...)
}

func (scheduler *SimpleScheduler) execute(ctx context.Context, task Task, interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			task(ctx)
		case <-ctx.Done():
			scheduler.wg.Done()
			return
		}
	}
}
