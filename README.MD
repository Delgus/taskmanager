
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/Delgus/taskmanager)
[![codecov](https://codecov.io/gh/Delgus/taskmanager/branch/master/graph/badge.svg)](https://codecov.io/gh/Delgus/taskmanager)
[![Go Report Card](https://goreportcard.com/badge/github.com/delgus/taskmanager)](https://goreportcard.com/report/github.com/delgus/taskmanager)
[![Release](https://img.shields.io/github/v/release/delgus/taskmanager)](https://img.shields.io/github/v/release/delgus/taskmanager)
[![Go version](https://img.shields.io/github/go-mod/go-version/delgus/taskmanager)](https://img.shields.io/github/go-mod/go-version/delgus/taskmanager)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/delgus/taskmanager/master/LICENSE)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager?ref=badge_shield)


### TASK MANAGER

This library implemented a system of task queues (library).
TaskManager should resolve tasks, tasks can be of various types.
Tasks can have different priority levels:

 - lowest
 - low
 - middle
 - high
 - highest

Tasks have events:

 - BeforeExecEvent
 - AfterExecEvent
 - FailedEvent
 
#### Create task with priority

```go
job := func() error {
	fmt.Println("doing work!")
	return nil
}
task := taskmanager.NewTask(taskmanager.HighestPriority, job)
```

#### add events to tasks
```go
task.OnEvent(taskmanager.BeforeExecEvent, func() {
	fmt.Println("i before execution!")
})
```

#### Create queue
```go
queue := NewMemoryQueue()
```

#### add task to queue
```go
queue.AddTask(task)
```

#### get task from queue
```go
queue.GetTask()
```

#### worker pool
```go
workerPool := taskmanager.NewWorkerPool(queue, 10, time.Millisecond*50, logger)
```



### License  

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager?ref=badge_large)
=======
