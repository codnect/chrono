package chrono

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDefaultTaskExecutor(t *testing.T) {
	executor := NewDefaultTaskExecutor()

	var counter int32

	task, err := executor.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 1*time.Second)

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleTaskExecutor_WithoutTaskRunner(t *testing.T) {
	executor := NewSimpleTaskExecutor(nil)

	var counter int32

	task, err := executor.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 1*time.Second)

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleTaskExecutor_Schedule_OneShotTask(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task, err := executor.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 1*time.Second)

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleTaskExecutor_ScheduleWithFixedDelay(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task, err := executor.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 0, 200*time.Millisecond)

	assert.Nil(t, err)

	<-time.After(1*time.Second + 500*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 3,
		"number of scheduled task execution must be between 1 and 3, actual: %d", counter)
}

func TestSimpleTaskExecutor_ScheduleWithFixedDelayWithInitialDelay(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task, err := executor.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 1*time.Second, 200*time.Millisecond)

	assert.Nil(t, err)

	<-time.After(2*time.Second + 500*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 3,
		"number of scheduled task execution must be between 1 and 3, actual: %d", counter)
}

func TestSimpleTaskExecutor_ScheduleAtFixedRate(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task, err := executor.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 0, 200*time.Millisecond)

	assert.Nil(t, err)

	<-time.After(2*time.Second - 50*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 10,
		"number of scheduled task execution must be between 5 and 10, actual: %d", counter)
}

func TestSimpleTaskExecutor_ScheduleAtFixedRateWithInitialDelay(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task, err := executor.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 1*time.Second, 200*time.Millisecond)

	assert.Nil(t, err)

	<-time.After(3*time.Second - 50*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 5 && counter <= 10,
		"number of scheduled task execution must be between 5 and 10, actual: %d", counter)
}

func TestSimpleTaskExecutor_Shutdown(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	executor.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 1*time.Second, 200*time.Millisecond)

	<-time.After(2 * time.Second)
	executor.Shutdown()

	expected := counter
	<-time.After(3 * time.Second)

	assert.True(t, executor.IsShutdown())
	assert.Equal(t, expected, counter,
		"after shutdown, previously scheduled tasks should not be rescheduled", counter)
}

func TestSimpleTaskExecutor_NoNewTaskShouldBeAccepted_AfterShutdown(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())
	executor.Shutdown()

	var err error
	_, err = executor.Schedule(func(ctx context.Context) {
	}, 1*time.Second)

	assert.NotNil(t, err)

	_, err = executor.ScheduleWithFixedDelay(func(ctx context.Context) {
	}, 1*time.Second, 1*time.Second)

	assert.NotNil(t, err)

	_, err = executor.ScheduleAtFixedRate(func(ctx context.Context) {
	}, 1*time.Second, 200*time.Millisecond)
	assert.NotNil(t, err)
}

func TestSimpleTaskExecutor_Schedule_MultiTasks(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())

	var task1Counter int32
	var task2Counter int32
	var task3Counter int32

	task1, err := executor.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&task1Counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 1*time.Second, 200*time.Millisecond)

	assert.Nil(t, err)

	task2, err := executor.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&task2Counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 0, 200*time.Millisecond)

	assert.Nil(t, err)

	task3, err := executor.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&task3Counter, 1)
	}, 0, 200*time.Millisecond)

	assert.Nil(t, err)

	<-time.After(2*time.Second - 50*time.Millisecond)

	task1.Cancel()
	task2.Cancel()
	task3.Cancel()

	assert.True(t, task1Counter >= 5 && task1Counter <= 10,
		"number of scheduled task 1 execution must be between 5 and 10, actual: %d", task1Counter)

	assert.True(t, task2Counter >= 1 && task2Counter <= 3,
		"number of scheduled task 2 execution must be between 1 and 3, actual: %d", task2Counter)

	assert.True(t, task3Counter >= 1 && task3Counter <= 10,
		"number of scheduled task execution must be between 5 and 10, actual: %d", task3Counter)
}

func TestSimpleTaskExecutor_ScheduleWithNilTask(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())

	var task ScheduledTask
	var err error

	task, err = executor.Schedule(nil, 1*time.Second)
	assert.Nil(t, task)
	assert.NotNil(t, err)

	task, err = executor.ScheduleWithFixedDelay(nil, 1*time.Second, 1*time.Second)
	assert.Nil(t, task)
	assert.NotNil(t, err)

	task, err = executor.ScheduleAtFixedRate(nil, 1*time.Second, 1*time.Second)
	assert.Nil(t, task)
	assert.NotNil(t, err)
}

func TestSimpleTaskExecutor_Shutdown_TerminatedExecutor(t *testing.T) {
	executor := NewSimpleTaskExecutor(NewDefaultTaskRunner())
	executor.Shutdown()

	assert.Panics(t, func() {
		executor.Shutdown()
	})
}
