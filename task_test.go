package chrono

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestNewSchedulerTask(t *testing.T) {
	assert.Panics(t, func() {
		NewSchedulerTask(nil)
	})
}

func TestNewSchedulerTask_WithLocation(t *testing.T) {
	assert.NotPanics(t, func() {
		NewSchedulerTask(func(ctx context.Context) {

		}, WithLocation("Europe/Istanbul"))
	})
}

func TestNewSchedulerTask_WithInvalidLocation(t *testing.T) {
	assert.Panics(t, func() {
		NewSchedulerTask(func(ctx context.Context) {

		}, WithLocation("Europe"))
	})
}

func TestNewScheduledRunnableTask(t *testing.T) {
	task := NewScheduledRunnableTask(0, func(ctx context.Context) {

	}, time.Now(), -1, true)

	assert.Equal(t, task.period, 0*time.Second)

	assert.Panics(t, func() {
		NewScheduledRunnableTask(0, nil, time.Now(), -1, true)
	})
}

func TestNewTriggerTask(t *testing.T) {
	assert.Panics(t, func() {
		NewTriggerTask(nil, NewDefaultScheduledExecutor(), NewCronTrigger("* * * * * *", time.Local))
	})

	assert.Panics(t, func() {
		NewTriggerTask(func(ctx context.Context) {

		}, nil, NewCronTrigger("* * * * * *", time.Local))
	})

	assert.Panics(t, func() {
		NewTriggerTask(func(ctx context.Context) {

		}, NewDefaultScheduledExecutor(), nil)
	})
}

type zeroTrigger struct {
}

func (trigger *zeroTrigger) NextExecutionTime(ctx TriggerContext) time.Time {
	return time.Time{}
}

func TestTriggerTask_Schedule(t *testing.T) {
	task := NewTriggerTask(func(ctx context.Context) {}, NewDefaultScheduledExecutor(), &zeroTrigger{})
	_, err := task.Schedule()
	assert.NotNil(t, err)
}

type scheduledExecutorMock struct {
	mock.Mock
}

func (executor scheduledExecutorMock) Schedule(task Task, delay time.Duration) (ScheduledTask, error) {
	result := executor.Called(task, delay)
	return result.Get(0).(ScheduledTask), result.Error(1)
}

func (executor scheduledExecutorMock) ScheduleWithFixedDelay(task Task, initialDelay time.Duration, delay time.Duration) (ScheduledTask, error) {
	result := executor.Called(task, initialDelay, delay)
	return result.Get(0).(ScheduledTask), result.Error(1)
}

func (executor scheduledExecutorMock) ScheduleAtFixedRate(task Task, initialDelay time.Duration, period time.Duration) (ScheduledTask, error) {
	result := executor.Called(task, initialDelay, period)
	return result.Get(0).(ScheduledTask), result.Error(1)
}

func (executor scheduledExecutorMock) IsShutdown() bool {
	result := executor.Called()
	return result.Bool(0)
}

func (executor scheduledExecutorMock) Shutdown() chan bool {
	result := executor.Called()
	return result.Get(0).(chan bool)
}

func TestTriggerTask_ScheduleWithError(t *testing.T) {
	scheduledExecutorMock := &scheduledExecutorMock{}

	scheduledExecutorMock.On("Schedule", mock.AnythingOfType("Task"), mock.AnythingOfType("time.Duration")).
		Return((*ScheduledRunnableTask)(nil), errors.New("test error"))

	task := NewTriggerTask(func(ctx context.Context) {}, scheduledExecutorMock, NewCronTrigger("* * * * * * ", time.Local))
	_, err := task.Schedule()

	assert.NotNil(t, err)
}
