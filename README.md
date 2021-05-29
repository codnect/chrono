![chrono](https://user-images.githubusercontent.com/5354910/118358070-739adb00-b57d-11eb-989b-68baf83f9584.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/procyon-projects/chrono)](https://goreportcard.com/report/github.com/procyon-projects/chrono)
[![Build Status](https://travis-ci.com/procyon-projects/chrono.svg?branch=main)](https://travis-ci.com/procyon-projects/chrono)
[![codecov](https://codecov.io/gh/procyon-projects/chrono/branch/main/graph/badge.svg?token=OREV0YI8VU)](https://codecov.io/gh/procyon-projects/chrono)

Chrono is a scheduler library which lets you run your task and code periodically

## Schedule Task With Fixed Delay
Scheduling a task at a Fixed Delay can be done with the help of the **ScheduleWithFixedDelay** method.

## Schedule Task at a Fixed Rate
Scheduling a task at a Fixed Rate can be done with the help of the **ScheduleAtFixedRate** method.

### Scheduling the Task at a Fixed Rate

Let's schedule a task to run at a fixed rate of seconds.

```go
taskScheduler := chrono.NewDefaultTaskScheduler()

task, err := taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
	log.Print("Fixed Rate of 5 seconds")
}, 5 * time.Second)

if err == nil {
	log.Print("Task has been scheduled")
}
```

The next task will run always after 5 seconds no matter the status of previous task, which may be still running. So even if the previous task isn't done, the next task will run.


### Scheduling the Task at a Fixed Rate From a Given Date
Let's schedule a task to run at a fixed rate from a given date.

```go
taskScheduler := chrono.NewDefaultTaskScheduler()

now := time.Now()

task, err := taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
	log.Print("Fixed Rate of 5 seconds")
}, 5 * time.Second, WithStartTime(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second() + 2))

if err == nil {
	log.Print("Task has been scheduled")
}
```

The task will be executed the first time 2 seconds after the current time, and it will continue to be executed according to the given fixed rate.

## Schedule Task With Cron Expression
Sometimes Fixed Delay and Fixed Rate are not enough, and we need the flexibility of a cron expression to schedule our tasks. With the help of the provided **ScheduleWithCron** method, we can schedule a task based on a cron expression.

```go
taskScheduler := chrono.NewDefaultTaskScheduler()

task, err := taskScheduler.ScheduleWithCron(func(ctx context.Context) {
	log.Print("Scheduled Task With Cron")
}, "0 45 18 10 * *")

if err == nil {
	log.Print("Task has been scheduled")
}
```

In this case, we're scheduling a task to be executed at 18:45  on the 10th day of every month

By default, the local time is used for the cron expression. However, we can use the **WithLocation** option to change this.

```go
task, err := taskScheduler.ScheduleWithCron(func(ctx context.Context) {
	log.Print("Scheduled Task With Cron")
}, "0 45 18 10 * *", WithLocation("America/New_York"))
```

In the above example, Task will be scheduled to be executed at 18:45 on the 10th day of every month in America/New York time.
