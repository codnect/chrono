package chrono

import (
	"context"
	"github.com/stretchr/testify/assert"
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
