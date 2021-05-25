package chrono

import "time"

type TriggerContext interface {
	LastCompletionTime() time.Time
	LastExecutionTime() time.Time
	LastScheduledExecutionTime() time.Time
}

type SimpleTriggerContext struct {
	lastCompletionTime         time.Time
	lastExecutionTime          time.Time
	lastScheduledExecutionTime time.Time
}

func NewSimpleTriggerContext() *SimpleTriggerContext {
	return &SimpleTriggerContext{}
}

func (ctx *SimpleTriggerContext) Update(lastCompletionTime time.Time, lastExecutionTime time.Time, lastScheduledExecutionTime time.Time) {
	ctx.lastCompletionTime = lastCompletionTime
	ctx.lastExecutionTime = lastExecutionTime
	ctx.lastScheduledExecutionTime = lastScheduledExecutionTime
}

func (ctx *SimpleTriggerContext) LastCompletionTime() time.Time {
	return ctx.lastCompletionTime
}

func (ctx *SimpleTriggerContext) LastExecutionTime() time.Time {
	return ctx.lastExecutionTime
}

func (ctx *SimpleTriggerContext) LastScheduledExecutionTime() time.Time {
	return ctx.lastScheduledExecutionTime
}

type Trigger interface {
	NextExecutionTime(ctx TriggerContext) time.Time
}

type CronTrigger struct {
	cronExpression *CronExpression
	location       *time.Location
}

func NewCronTrigger(expression string, location *time.Location) *CronTrigger {
	cron, err := ParseCronExpression(expression)

	if err != nil {
		panic(err)
	}

	trigger := &CronTrigger{
		cron,
		time.Local,
	}

	if location != nil {
		trigger.location = location
	}

	return trigger
}

func (trigger *CronTrigger) NextExecutionTime(ctx TriggerContext) time.Time {
	now := time.Now()
	lastCompletion := ctx.LastCompletionTime()

	if !lastCompletion.IsZero() {

		lastExecution := ctx.LastScheduledExecutionTime()

		if !lastExecution.IsZero() && now.Before(lastExecution) {
			now = lastExecution
		}

	}

	originalLocation := now.Location()
	next := trigger.cronExpression.NextTime(now.In(trigger.location))
	return next.In(originalLocation)
}
