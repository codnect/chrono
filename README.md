![Chrono Logo](https://user-images.githubusercontent.com/5354910/196008920-1ca88967-3d7d-449c-b165-fe38c5e1fb57.png)
# Chrono
[![Go Report Card](https://goreportcard.com/badge/github.com/procyon-projects/chrono)](https://goreportcard.com/report/github.com/procyon-projects/chrono)
[![CircleCI](https://dl.circleci.com/status-badge/img/gh/procyon-projects/chrono/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/procyon-projects/chrono/tree/main)
[![codecov](https://codecov.io/gh/procyon-projects/chrono/branch/main/graph/badge.svg?token=OREV0YI8VU)](https://codecov.io/gh/procyon-projects/chrono)

Chrono is a scheduler library that lets you run your tasks and code periodically. It provides different scheduling functionalities to make it easier to create a scheduling task.

## Installation

```shell
go get github.com/procyon-projects/chrono@latest
```

## Scheduling a One-Shot Task
The Schedule method helps us schedule the task to run once at the specified time. In the following example, the task will first be executed 1 second after the current time.
**WithTime** option is used to specify the execution time.

```go
taskScheduler := chrono.NewDefaultTaskScheduler()
now := time.Now()
startTime := now.Add(time.Second * 1)

task, err := taskScheduler.Schedule(func(ctx context.Context) {
	log.Print("One-Shot Task")
}, chrono.WithTime(startTime))

if err == nil {
	log.Print("Task has been scheduled successfully.")
}
```

Also, **WithStartTime** option can be used to specify the execution time. **But It's deprecated.**

```go
taskScheduler := chrono.NewDefaultTaskScheduler()

task, err := taskScheduler.Schedule(func(ctx context.Context) {
	log.Print("One-Shot Task")
}, chrono.WithStartTime(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1))

if err == nil {
	log.Print("Task has been scheduled successfully.")
}
```

## Scheduling a Task with Fixed Delay
Let's schedule a task to run with a fixed delay between the finish time of the last execution of the task and the start time of the next execution of the task.
The fixed delay counts the delay after the completion of the last execution.

```go
taskScheduler := chrono.NewDefaultTaskScheduler()

task, err := taskScheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
	log.Print("Fixed Delay Task")
	time.Sleep(3 * time.Second)
}, 5 * time.Second)

if err == nil {
	log.Print("Task has been scheduled successfully.")
}
```

Since the task itself takes 3 seconds to complete and we have specified a delay of 5 seconds between the finish time of the last execution of the task and the start time of the next execution of the task, there will be a delay of 8 seconds between each execution.

**WithStartTime** and **WithLocation** options can be combined with this.

## Schedule Task at a Fixed Rate
Let's schedule a task to run at a fixed rate of seconds.

```go
taskScheduler := chrono.NewDefaultTaskScheduler()

task, err := taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
	log.Print("Fixed Rate of 5 seconds")
}, 5 * time.Second)

if err == nil {
	log.Print("Task has been scheduled successfully.")
}
```

The next task will run always after 5 seconds no matter the status of the previous task, which may be still running. So even if the previous task isn't done, the next task will run.
We can also use the **WithStartTime** option to specify the desired first execution time of the task.

```go
now := time.Now()

task, err := taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
	log.Print("Fixed Rate of 5 seconds")
}, 5 * time.Second, chrono.WithStartTime(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second() + 2))
```

When we use this option, the task will run at the specified execution time and subsequently with the given period. In the above example, the task will first be executed 2 seconds after the current time.

We can also combine this option with **WithLocation** based on our requirements.

```go
now := time.Now()

task, err := taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
	log.Print("Fixed Rate of 5 seconds")
}, 5 * time.Second, chrono.WithStartTime(now.Year(), now.Month(), now.Day(), 18, 45, 0),
chrono.WithLocation("America/New_York"))
```

In the above example, the task will first be executed at 18:45 of the current date in America/New York time.
**If the start time is in the past, the task will be executed immediately.**

## Scheduling a Task using Cron Expression
Sometimes Fixed Rate and Fixed Delay can not fulfill your needs, and we need the flexibility of cron expressions to schedule the execution of your tasks. With the help of the provided **ScheduleWithCron method**, we can schedule a task based on a cron expression.

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
}, "0 45 18 10 * *", chrono.WithLocation("America/New_York"))
```

In the above example, Task will be scheduled to be executed at 18:45 on the 10th day of every month in America/New York time.

**WithStartTimeoption** cannot be used with **ScheduleWithCron**.

## Canceling a Scheduled Task
Schedule methods return an instance of type ScheduledTask, which allows us to cancel a task or to check if the task is canceled. The Cancel method cancels the scheduled task but running tasks won't be interrupted.


```go
taskScheduler := chrono.NewDefaultTaskScheduler()

task, err := taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
	log.Print("Fixed Rate of 5 seconds")
}, 5 * time.Second)

/* ... */
	
task.Cancel()
```

## Shutting Down a Scheduler
The **Shutdown()** method doesn't cause immediate shut down of the Scheduler and returns a channel. It will make the Scheduler stop accepting new tasks and shut down after all running tasks finish their current work.


```go
taskScheduler := chrono.NewDefaultTaskScheduler()

/* ... */

shutdownChannel := taskScheduler.Shutdown()
<- shutdownChannel
	
/* after all running task finished their works */
```

Stargazers
-----------
[![Stargazers repo roster for @procyon-projects/chrono](https://reporoster.com/stars/procyon-projects/chrono)](https://github.com/procyon-projects/chrono/stargazers)

Forkers
-----------
[![Forkers repo roster for @procyon-projects/chrono](https://reporoster.com/forks/procyon-projects/chrono)](https://github.com/procyon-projects/chrono/network/members)

# License
Chrono is released under MIT License.
