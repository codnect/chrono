package chrono

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultTaskScheduler(t *testing.T) {
	scheduler := NewDefaultTaskScheduler()

	var counter int32
	now := time.Now()

	task, err := scheduler.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, WithStartTime(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1))

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestDefaultTaskSchedulerWithTimeOption(t *testing.T) {
	scheduler := NewDefaultTaskScheduler()

	var counter int32
	now := time.Now()
	starTime := now.Add(time.Second * 1)

	task, err := scheduler.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, WithTime(starTime))

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleWithoutTask(t *testing.T) {
	scheduler := NewDefaultTaskScheduler()
	task, err := scheduler.Schedule(nil)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestSimpleTaskScheduler_ScheduleWithFixedDelayWithoutTask(t *testing.T) {
	scheduler := NewDefaultTaskScheduler()
	task, err := scheduler.ScheduleWithFixedDelay(nil, 2*time.Second)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestSimpleTaskScheduler_ScheduleAtFixedRateWithoutTask(t *testing.T) {
	scheduler := NewDefaultTaskScheduler()
	task, err := scheduler.ScheduleAtFixedRate(nil, 2*time.Second)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestSimpleTaskScheduler_ScheduleWithCronWithoutTask(t *testing.T) {
	scheduler := NewDefaultTaskScheduler()
	task, err := scheduler.ScheduleWithCron(nil, "* * * * * *")
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestSimpleTaskScheduler_ScheduleWithCronUsingInvalidCronExpresion(t *testing.T) {
	scheduler := NewDefaultTaskScheduler()
	task, err := scheduler.ScheduleWithCron(func(ctx context.Context) {

	}, "test * * * * *")
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestSimpleTaskScheduler_WithoutScheduledExecutor(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(nil)

	var counter int32
	now := time.Now()

	task, err := scheduler.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, WithStartTime(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1))

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleTaskScheduler_WithoutScheduledExecutorWithTimeOption(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(nil)

	var counter int32
	now := time.Now()
	startTime := now.Add(time.Second * 1)

	task, err := scheduler.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, WithTime(startTime))

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleOneShotTaskWithStartTimeOption(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32
	now := time.Now()

	task, err := scheduler.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, WithStartTime(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1))

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleOneShotTaskWithTimeOption(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32
	now := time.Now()
	startTime := now.Add(time.Second * 1)

	task, err := scheduler.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, WithTime(startTime))

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleWithFixedDelay(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32

	task, err := scheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 200*time.Millisecond)

	assert.Nil(t, err)

	<-time.After(1*time.Second + 500*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 3,
		"number of scheduled task execution must be between 1 and 3, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleWithFixedDelayWithStartTimeOption(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32
	now := time.Now()

	task, err := scheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 200*time.Millisecond, WithStartTime(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1))

	assert.Nil(t, err)

	<-time.After(2*time.Second + 500*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 3,
		"number of scheduled task execution must be between 1 and 3, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleWithFixedDelayWithTimeOption(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32
	now := time.Now()
	startTime := now.Add(time.Second * 1)

	task, err := scheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 200*time.Millisecond, WithTime(startTime))

	assert.Nil(t, err)

	<-time.After(2*time.Second + 500*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 3,
		"number of scheduled task execution must be between 1 and 3, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleAtFixedRate(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32

	task, err := scheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 200*time.Millisecond)

	assert.Nil(t, err)

	<-time.After(1*time.Second + 950*time.Microsecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 10,
		"number of scheduled task execution must be between 5 and 10, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleAtFixedRateWithStartTimeOption(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32
	now := time.Now()

	task, err := scheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 200*time.Millisecond, WithStartTime(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1))

	assert.Nil(t, err)

	<-time.After(3*time.Second - 50*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 5 && counter <= 10,
		"number of scheduled task execution must be between 5 and 10, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleAtFixedRateWithTimeOption(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32
	now := time.Now()
	startTime := now.Add(time.Second * 1)

	task, err := scheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 200*time.Millisecond, WithTime(startTime))

	assert.Nil(t, err)

	<-time.After(3*time.Second - 50*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 5 && counter <= 10,
		"number of scheduled task execution must be between 5 and 10, actual: %d", counter)
}

func TestSimpleTaskScheduler_ScheduleWithCron(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32

	task, err := scheduler.ScheduleWithCron(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, "0-59/2 * * * * *")

	assert.Nil(t, err)

	<-time.After(10 * time.Second)
	task.Cancel()
	assert.True(t, counter >= 5,
		"number of scheduled task execution must be at least 5, actual: %d", counter)
}

func TestSimpleTaskScheduler_Shutdown(t *testing.T) {
	scheduler := NewSimpleTaskScheduler(NewDefaultTaskExecutor())

	var counter int32

	_, err := scheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 1*time.Second)

	assert.Nil(t, err)

	<-time.After(2 * time.Second)
	scheduler.Shutdown()

	expected := counter
	<-time.After(3 * time.Second)

	assert.True(t, scheduler.IsShutdown())
	assert.Equal(t, expected, counter,
		"after shutdown, previously scheduled tasks should not be rescheduled", counter)
}
