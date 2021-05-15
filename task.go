package chrono

import (
	"context"
	"sort"
	"time"
)

type Task func(ctx context.Context)

type ScheduledTask interface {
	Cancel()
	NextExecutionTime() time.Time
}

type RunnableTask struct {
	id        int
	task      Task
	startTime time.Time
	location  *time.Location
}

func newRunnableTask(id int, task Task, options ...Option) *RunnableTask {
	runnableTask := &RunnableTask{
		id:   id,
		task: task,
	}

	for _, option := range options {
		option(runnableTask)
	}

	return runnableTask
}

func (task *RunnableTask) update() {

}

type Option func(task *RunnableTask)

func WithStartTime(startTime time.Time) Option {
	return func(task *RunnableTask) {
		task.startTime = startTime
	}
}

func WithLocation(location string) Option {
	return func(task *RunnableTask) {
		loadedLocation, err := time.LoadLocation(location)

		if err != nil {
			panic(err)
		}

		task.location = loadedLocation
	}
}

type OneShotTask struct {
	*RunnableTask
	delay time.Duration
}

func newOneShotTask(id int, task Task, options ...Option) *OneShotTask {
	return &OneShotTask{
		RunnableTask: newRunnableTask(id, task, options...),
	}
}

func (task *OneShotTask) Cancel() {

}

func (task *OneShotTask) NextExecutionTime() time.Time {
	return time.Time{}
}

type FixedDelayTask struct {
	*RunnableTask
	delay time.Duration
}

func newFixedDelayTask(id int, task Task, options ...Option) *FixedDelayTask {
	return &FixedDelayTask{
		RunnableTask: newRunnableTask(id, task, options...),
	}
}

func (task *FixedDelayTask) Cancel() {

}

func (task *FixedDelayTask) NextExecutionTime() time.Time {
	return time.Time{}
}

type FixedRateTask struct {
	*RunnableTask
	period time.Duration
}

func newFixedRateTask(id int, task Task, options ...Option) *FixedRateTask {
	return &FixedRateTask{
		RunnableTask: newRunnableTask(id, task, options...),
	}
}

func (task *FixedRateTask) Cancel() {

}

func (task *FixedRateTask) NextExecutionTime() time.Time {
	return time.Time{}
}

type CronTask struct {
	*RunnableTask
}

func newCronTask(id int, task Task, options ...Option) *CronTask {
	return &CronTask{
		RunnableTask: newRunnableTask(id, task, options...),
	}
}

func (task *CronTask) Cancel() {

}

func (task *CronTask) NextExecutionTime() time.Time {
	return time.Time{}
}

func (task *RunnableTask) updateNextExecutionTime(t time.Time) {

}

func (task *RunnableTask) execute(ctx context.Context) {

}

type Tasks []*RunnableTask

func (tasks Tasks) IsEmpty() bool {
	return tasks.Len() == 0
}

func (tasks Tasks) UpdateNextExecutionTimes(t time.Time) {

	for _, task := range tasks {
		task.updateNextExecutionTime(t)
	}

}

func (tasks Tasks) SortByNextExecutionTime() {
	sort.Sort(tasks)
}

func (tasks Tasks) Len() int {
	return len(tasks)
}

func (tasks Tasks) Swap(i, j int) {
	tasks[i], tasks[j] = tasks[j], tasks[i]
}

func (tasks Tasks) Less(i, j int) bool {
	return false
}
