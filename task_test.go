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
	_, err := CreateSchedulerTask(nil)
	assert.Error(t, err)
}

func TestNewSchedulerTask_WithLocation(t *testing.T) {
	_, err := CreateSchedulerTask(func(ctx context.Context) {

	}, WithLocation("Europe/Istanbul"))
	assert.Nil(t, err)
}

func TestNewSchedulerTask_WithInvalidLocation(t *testing.T) {
	_, err := CreateSchedulerTask(func(ctx context.Context) {

	}, WithLocation("Europe"))
	assert.Error(t, err)
}

func TestNewScheduledRunnableTask(t *testing.T) {
	task, _ := CreateScheduledRunnableTask(0, func(ctx context.Context) {

	}, time.Now(), -1, true)

	assert.Equal(t, task.period, 0*time.Second)

	_, err := CreateScheduledRunnableTask(0, nil, time.Now(), -1, true)
	assert.Error(t, err)
}

func TestNewTriggerTask(t *testing.T) {
	trigger, err := CreateCronTrigger("* * * * * *", time.Local)
	assert.Nil(t, err)

	_, err = CreateTriggerTask(nil, NewDefaultTaskExecutor(), trigger)
	assert.Error(t, err)

	_, err = CreateTriggerTask(func(ctx context.Context) {

	}, nil, trigger)
	assert.Error(t, err)

	_, err = CreateTriggerTask(func(ctx context.Context) {

	}, NewDefaultTaskExecutor(), nil)
}

type zeroTrigger struct {
}

func (trigger *zeroTrigger) NextExecutionTime(ctx TriggerContext) time.Time {
	return time.Time{}
}

func TestTriggerTask_Schedule(t *testing.T) {
	task, _ := CreateTriggerTask(func(ctx context.Context) {}, NewDefaultTaskExecutor(), &zeroTrigger{})
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

	trigger, err := CreateCronTrigger("* * * * * *", time.Local)
	assert.Nil(t, err)

	task, _ := CreateTriggerTask(func(ctx context.Context) {}, scheduledExecutorMock, trigger)
	_, err = task.Schedule()

	assert.NotNil(t, err)
}
