package chrono

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultScheduler(t *testing.T) {
	scheduler := NewDefaultScheduler()

	var counter int32
	now := time.Now()

	task := scheduler.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, WithStartTime(Time(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1, now.Nanosecond())))

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleScheduler_WithoutScheduledExecutor(t *testing.T) {
	scheduler := NewSimpleScheduler(nil)

	var counter int32
	now := time.Now()

	task := scheduler.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, WithStartTime(Time(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1, now.Nanosecond())))

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleScheduler_Schedule_OneShotTask(t *testing.T) {
	scheduler := NewSimpleScheduler(NewDefaultScheduledExecutor())

	var counter int32
	now := time.Now()

	task := scheduler.Schedule(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, WithStartTime(Time(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1, now.Nanosecond())))

	<-time.After(2 * time.Second)
	assert.True(t, task.IsCancelled(), "scheduled task must have been cancelled")
	assert.True(t, counter == 1,
		"number of scheduled task execution must be 1, actual: %d", counter)
}

func TestSimpleScheduler_ScheduleWithFixedDelay(t *testing.T) {
	scheduler := NewSimpleScheduler(NewDefaultScheduledExecutor())

	var counter int32

	task := scheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 200*time.Millisecond)

	<-time.After(1*time.Second + 500*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 3,
		"number of scheduled task execution must be between 1 and 2, actual: %d", counter)
}

func TestSimpleScheduler_ScheduleWithFixedDelayWithStartTimeOption(t *testing.T) {
	scheduler := NewSimpleScheduler(NewDefaultScheduledExecutor())

	var counter int32
	now := time.Now()

	task := scheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 200*time.Millisecond, WithStartTime(
		Time(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1, now.Nanosecond())))

	<-time.After(2*time.Second + 500*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 3,
		"number of scheduled task execution must be between 1 and 2, actual: %d", counter)
}

func TestSimpleScheduler_ScheduleAtFixedRate(t *testing.T) {
	scheduler := NewSimpleScheduler(NewDefaultScheduledExecutor())

	var counter int32

	task := scheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}, 200*time.Millisecond)

	<-time.After(2 * time.Second)
	task.Cancel()
	assert.True(t, counter >= 1 && counter <= 10,
		"number of scheduled task execution must be between 5 and 10, actual: %d", counter)
}

func TestSimpleScheduler_ScheduleAtFixedRateWithStartTimeOption(t *testing.T) {
	scheduler := NewSimpleScheduler(NewDefaultScheduledExecutor())

	var counter int32
	now := time.Now()

	task := scheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, 200*time.Millisecond, WithStartTime(
		Time(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1, now.Nanosecond())))

	<-time.After(3*time.Second + 200*time.Millisecond)
	task.Cancel()
	assert.True(t, counter >= 5 && counter <= 10,
		"number of scheduled task execution must be between 5 and 10")
}

func TestSimpleScheduler_ScheduleWithCron(t *testing.T) {
	scheduler := NewSimpleScheduler(NewDefaultScheduledExecutor())

	var counter int32

	task := scheduler.ScheduleWithCron(func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
		<-time.After(500 * time.Millisecond)
	}, "0-59/2 * * * * *")

	<-time.After(10 * time.Second)
	task.Cancel()
	assert.True(t, counter >= 5,
		"number of scheduled task execution must be at least 5")
}
