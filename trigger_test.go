package chrono

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSimpleTriggerContext(t *testing.T) {
	ctx := NewSimpleTriggerContext()
	now := time.Now()
	ctx.lastExecutionTime = now
	ctx.lastTriggeredExecutionTime = now
	ctx.lastCompletionTime = now

	assert.Equal(t, now, ctx.LastExecutionTime())
	assert.Equal(t, now, ctx.LastCompletionTime())
	assert.Equal(t, now, ctx.LastTriggeredExecutionTime())
}

func TestNewCronTrigger(t *testing.T) {
	trigger, err := CreateCronTrigger("", time.Local)
	assert.Error(t, err)
	assert.Nil(t, trigger)
}
