[![Sourcegraph](https://sourcegraph.com/github.com/Delgus/taskmanager/-/badge.svg)](https://sourcegraph.com/github.com/delgus/taskmanager?badge)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/Delgus/taskmanager)
[![codecov](https://codecov.io/gh/Delgus/taskmanager/branch/master/graph/badge.svg)](https://codecov.io/gh/Delgus/taskmanager)
[![Go Report Card](https://goreportcard.com/badge/github.com/delgus/taskmanager)](https://goreportcard.com/report/github.com/delgus/taskmanager)
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
 
### Examples in playground

simple queue  
[https://play.golang.org/p/6kcB9-jw-wU](https://play.golang.org/p/6kcB9-jw-wU)

worker
[https://play.golang.org/p/nXiMzS8mOe4](https://play.golang.org/p/nXiMzS8mOe4)



### License  

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FDelgus%2Ftaskmanager?ref=badge_large)
=======
