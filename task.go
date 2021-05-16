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
	id             int
	task           Task
	startTime      time.Time
	location       *time.Location
	trigger        Trigger
	triggerContext *SimpleTriggerContext
}

func NewRunnableTask(id int, task Task, options ...Option) RunnableTask {
	runnableTask := &RunnableTask{
		id:        id,
		task:      task,
		startTime: time.Time{},
		location:  time.Local,
	}

	for _, option := range options {
		option(runnableTask)
	}

	return *runnableTask
}

func (task *RunnableTask) Cancel() time.Time {
	return time.Time{}
}

func (task *RunnableTask) NextExecutionTime() time.Time {
	return task.trigger.NextExecutionTime(task.triggerContext)
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

type Tasks []*RunnableTask

func (tasks Tasks) IsEmpty() bool {
	return tasks.Len() == 0
}

func (tasks Tasks) UpdateNextExecutionTimes(t time.Time) {

	//for _, task := range tasks {
	//task.updateNextExecutionTime(t)
	//}

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
