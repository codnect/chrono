package scheduler

import (
	"context"
	"sync"
	"time"
)

type Option func(task *ScheduledTask)

func WithAsync() Option {
	return func(task *ScheduledTask) {

		if task.fixedRate == -1 {
			panic("you can only run a task with a fixed rate as async")
		}

		task.isAsync = true
	}
}

func WithInitialDelay(initialDelay time.Duration) Option {
	return func(task *ScheduledTask) {

		if task.isAsync {
			panic("async task does not support this option")
		}

		if initialDelay > 0 {
			task.initialDelay = initialDelay
		}

	}
}

func WithCron(expression string) Option {
	return func(task *ScheduledTask) {
		_, err := cronParser.Parse(expression)

		if err != nil {
			panic(err)
		}

	}
}

func WithFixedDelay(delay time.Duration) Option {
	return func(task *ScheduledTask) {

		if task.isAsync {
			panic("async task does not support this option")
		}

		if delay >= 0 {
			task.fixedRate = -1
			task.fixedDelay = delay
		}

	}
}

func WithFixedRate(period time.Duration) Option {
	return func(task *ScheduledTask) {

		if period >= 0 {
			task.fixedDelay = -1
			task.fixedRate = period
		}

	}
}

func WithLocation(location string) Option {
	return func(task *ScheduledTask) {
		loadedLocation, err := time.LoadLocation(location)

		if err != nil {
			panic(err)
		}

		task.location = loadedLocation
	}
}

type Scheduler interface {
	Schedule(ctx context.Context, task Task, options ...Option)
	Terminate()
}

type SimpleScheduler struct {
	wg              *sync.WaitGroup
	cancelFunctions []context.CancelFunc
}

func NewScheduler() Scheduler {
	return &SimpleScheduler{
		wg:              &sync.WaitGroup{},
		cancelFunctions: make([]context.CancelFunc, 0),
	}
}

func (scheduler *SimpleScheduler) Schedule(ctx context.Context, task Task, options ...Option) {
	scheduledTask := NewScheduledTask(task, options...)
	scheduledTask.Execute(nil)

	ctxWithCancel, cancel := context.WithCancel(ctx)
	scheduler.wg.Add(1)

	scheduler.cancelFunctions = append(scheduler.cancelFunctions, cancel)
	go scheduler.execute(ctxWithCancel, task, 1*time.Second)
}

func (scheduler *SimpleScheduler) Terminate() {
	for _, cancelFunction := range scheduler.cancelFunctions {
		cancelFunction()
	}

	scheduler.wg.Wait()
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
