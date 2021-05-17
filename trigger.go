package chrono

import "time"

type TriggerContext interface {
	GetTime() time.Time
	LastCompletionTime() time.Time
	LastExecutionTime() time.Time
	LastScheduledExecutionTime() time.Time
}

type SimpleTriggerContext struct {
	clockTime                  time.Time
	lastCompletionTime         time.Time
	lastExecutionTime          time.Time
	lastScheduledExecutionTime time.Time
}

func NewSimpleTriggerContext() *SimpleTriggerContext {
	return &SimpleTriggerContext{
		clockTime: time.Now(),
	}
}

func (ctx *SimpleTriggerContext) update(lastCompletionTime time.Time, lastExecutionTime time.Time, lastScheduledExecutionTime time.Time) {
	ctx.lastCompletionTime = lastCompletionTime
	ctx.lastExecutionTime = lastExecutionTime
	ctx.lastScheduledExecutionTime = lastScheduledExecutionTime
}

func (ctx *SimpleTriggerContext) GetTime() time.Time {
	return ctx.clockTime
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

type PeriodicTrigger struct {
	initialDelay time.Duration
	period       time.Duration
	fixedRate    bool
	location     *time.Location
}

func NewPeriodicTrigger(period, initialDelay time.Duration, fixedRate bool) *PeriodicTrigger {

	if initialDelay < 0 {
		initialDelay = 0
	}

	if period <= 0 {
		panic("period must be positive")
	}

	return &PeriodicTrigger{
		initialDelay: initialDelay,
		period:       period,
		fixedRate:    fixedRate,
	}
}

func (trigger *PeriodicTrigger) GetInitialDelay() time.Duration {
	return trigger.initialDelay
}

func (trigger *PeriodicTrigger) GetPeriod() time.Duration {
	return trigger.period
}

func (trigger *PeriodicTrigger) IsFixedRate() bool {
	return trigger.fixedRate
}

func (trigger *PeriodicTrigger) NextExecutionTime(ctx TriggerContext) time.Time {
	lastCompletion := ctx.LastCompletionTime()
	lastExecution := ctx.LastScheduledExecutionTime()

	if lastCompletion.IsZero() || lastExecution.IsZero() {
		return time.Now().Add(trigger.initialDelay)
	}

	if !trigger.fixedRate {
		return lastCompletion.Add(trigger.period)
	}

	return lastExecution.Add(trigger.period)
}

type CronTrigger struct {
	cronExpression *CronExpression
	location       *time.Location
}

func NewCronTrigger(expression string, location *time.Location) *CronTrigger {
	cron, err := NewCronParser().Parse(expression)

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
