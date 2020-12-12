
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/Delgus/taskmanager)
[![codecov](https://codecov.io/gh/Delgus/taskmanager/branch/master/graph/badge.svg)](https://codecov.io/gh/Delgus/taskmanager)
[![Go Report Card](https://goreportcard.com/badge/github.com/delgus/taskmanager)](https://goreportcard.com/report/github.com/delgus/taskmanager)
[![Release](https://img.shields.io/github/v/release/delgus/taskmanager)](https://img.shields.io/github/v/release/delgus/taskmanager)
[![Go version](https://img.shields.io/github/go-mod/go-version/delgus/taskmanager)](https://img.shields.io/github/go-mod/go-version/delgus/taskmanager)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/delgus/taskmanager/master/LICENSE)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager?ref=badge_shield)


# TASK MANAGER

This library implemented a system of task queues (library).
TaskManager should resolve tasks, tasks can be of various types.
Tasks can have different priority levels:

 - lowest
 - low
 - middle
 - high
 - highest
 
### Create task with priority

```go
task := taskmanager.NewTask(taskmanager.HighestPriority, func() error {
	fmt.Println("doing work!")
	return nil
})
```

### Create queue
```go
queue := taskmanager.NewQueue()
```

### Add task to queue
```go
queue.AddTask(task)
```

### Create worker pool
```go
// without getting errors from tasks
wp1 := taskmanager.NewWorkerPool(queue, false)

// get errors from task
wp2 := taskmanager.NewWorkerPool(q, true)
go func() {
	for err := range wp.Errors {
		log.Println("err:", err)
	}
}()
```

### Auto-Scale Worker Pool Options

##### SetPollTaskTicker(ticker *time.Ticker)

Set duration for polling tasks. 
default - 200ms

##### SetScheduleTicker(ticker *time.Ticker) 

Set duration for check count of workers for adding new or removing.
default - 1s

##### SetMinWorkers(count uint32)

Set min count running workers in pool.
default - 2

##### SetMaxWorkers(count uint32) 

Set max count running workers in pool
default - 25

##### SetMinWorkingPercent(per uint32) 

Set min percent running workers. default - 50
Example, if 4/10 workers running that 40% < 50% and one of workers been remove from pool

##### SetMaxWorkingPercent(per uint32)

Set max percent running workers. default - 99


### License  

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager?ref=badge_large)
=======
