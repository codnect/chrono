package chrono

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDefaultScheduledExecutor(t *testing.T) {
	executor := NewDefaultScheduledExecutor()

	var counter int32

	task := executor.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 1*time.Second)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestScheduledTaskExecutor_WithoutTaskRunner(t *testing.T) {
	executor := NewScheduledTaskExecutor(nil)

	var counter int32

	task := executor.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 1*time.Second)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestScheduledTaskExecutor_Schedule_OneShotTask(t *testing.T) {
	executor := NewScheduledTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task := executor.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 1*time.Second)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestScheduledTaskExecutor_ScheduleWithFixedDelay(t *testing.T) {
	executor := NewScheduledTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task := executor.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 0, 200*time.Millisecond)

	<-time.After(1*time.Second + 500*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 3,
		"number of scheduled task execution must be between 1 and 3, actual: %d", counter)
}

func TestScheduledTaskExecutor_ScheduleWithFixedDelayWithInitialDelay(t *testing.T) {
	executor := NewScheduledTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task := executor.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 1*time.Second, 200*time.Millisecond)

	<-time.After(2*time.Second + 500*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 3,
		"number of scheduled task execution must be between 1 and 3, actual: %d", counter)
}

func TestScheduledTaskExecutor_ScheduleAtFixedRate(t *testing.T) {
	executor := NewScheduledTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task := executor.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 0, 200*time.Millisecond)

	<-time.After(2*time.Second - 50*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 10,
		"number of scheduled task execution must be between 5 and 10, actual: %d", counter)
}

func TestScheduledTaskExecutor_ScheduleAtFixedRateWithInitialDelay(t *testing.T) {
	executor := NewScheduledTaskExecutor(NewDefaultTaskRunner())

	var counter int32

	task := executor.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 1*time.Second, 200*time.Millisecond)

	<-time.After(3*time.Second - 50*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 5 && counter <= 10,
		"number of scheduled task execution must be between 5 and 10, actual: %d", counter)
}

func TestScheduledTaskExecutor_Shutdown(t *testing.T) {
	executor := NewScheduledTaskExecutor(NewDefaultTaskRunner())

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

func TestScheduledTaskExecutor_NoNewTaskShouldBeAccepted_AfterShutdown(t *testing.T) {
	executor := NewScheduledTaskExecutor(NewDefaultTaskRunner())
	executor.Shutdown()

	assert.Panics(t, func() {
		executor.Schedule(func(ctx context.Context) {
		}, 1*time.Second)
	})

	assert.Panics(t, func() {
		executor.ScheduleWithFixedDelay(func(ctx context.Context) {
		}, 1*time.Second, 1*time.Second)
	})

	assert.Panics(t, func() {
		executor.ScheduleAtFixedRate(func(ctx context.Context) {
		}, 1*time.Second, 200*time.Millisecond)
	})
}
