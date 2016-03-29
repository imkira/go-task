# task

[![License](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://github.com/imkira/go-task/blob/master/LICENSE.txt)
[![GoDoc](https://godoc.org/github.com/imkira/go-task?status.svg)](https://godoc.org/github.com/imkira/go-task)
[![Build Status](http://img.shields.io/travis/imkira/go-task.svg?style=flat)](https://travis-ci.org/imkira/go-task)
[![Coverage](http://img.shields.io/codecov/c/github/imkira/go-task.svg?style=flat)](https://codecov.io/github/imkira/go-task)
[![codebeat badge](https://codebeat.co/badges/fd955959-f9c3-4732-b2f8-0f3a2a3d530e)](https://codebeat.co/projects/github-com-imkira-go-task)

task is a [Go](http://golang.org) utility package for simplifying execution of
tasks.

## Why

Go makes it easy to run your code concurrently. Many times you just want to run
a group of tasks in series, concurrently, or even compose them in order to
achieve more complex tasks. You also want to be able to cancel individual or a
group of tasks, wait for them to finish executing, or store and share data with
them.

This package draws inspiration from
[golang.org/x/net/context](https://godoc.org/golang.org/x/net/context) and
[github.com/gorilla/context](http://www.gorillatoolkit.org/pkg/context).

## Install

First, you need to install the package:

```go
go get -u github.com/imkira/go-task
```

Then, you need to include it in your source:

```go
import "github.com/imkira/go-task"
```

The package will be imported with ```task``` as name.

## Documentation

For advanced usage, make sure to check the
[available documentation here](http://godoc.org/github.com/imkira/go-task).

## How to Use

### Concepts

- Task: [Task](http://godoc.org/github.com/imkira/go-task#Task) is an interface
  of something you want to run. It can only be run once.
- Group: Group is a specialized type of of ```Task``` that run a group of tasks
  in a certain way. There are two types of Groups available:
  [ConcurrentGroup](http://godoc.org/github.com/imkira/go-task#ConcurrentGroup)
  and [SerialGroup](http://godoc.org/github.com/imkira/go-task#SerialGroup).
  They are for running tasks concurrently or in series, respectively.
- Context: [Context](http://godoc.org/github.com/imkira/go-task#Context) is an
  interface that allows you to store/retrieve data and share it with tasks.

### Example: Creating a Task

You can create a ```Task``` by adopting the
[Task](http://godoc.org/github.com/imkira/go-task#Task) interface.

The following are the recommended 2 ways to do it.

#### Using NewTaskWithFunc (easier)

```go
import "github.com/imkira/go-task"

func main() {
  // create the function to be run
  run := func(t task.Task, ctx task.Context) {
    // write your task below
    // ...
  }

  // create task passing it the function you want to run.
  t := task.NewTaskWithFunc(run)

  // run task (without context)
  t.Run(nil)
}
```

#### Embedding StandardTask in your struct

```go
import "github.com/imkira/go-task"

type fooTask struct {
  // embed StandardTask which includes the basic framework for Tasks.
  *task.StandardTask

  // add fields below
  // ...
}

// You just need to implement the Run function like this.
func (ft *fooTask) Run(ctx Context) {
  // In case you embed StandardTask, like in this example, the following
  // Start and Finish logic is required.
  if err := ft.Start(); err != nil {
    return
  }
  // defer here is absolutely required. If this task panics, the parent tasks
  // will wait for this task forever.
  defer ft.Finish()

  // finally, write your task below
  // ...
}

func main() {
  t := &fooTask{
    StandardTask: task.NewStandardTask(),
  }

  // run task (without context)
  t.Run(nil)
}
```

### Example: Creating a Task Group

#### Example: ConcurrentGroup

If you want to run tasks concurrently, you can do it like:

```go
func main() {
  // Create a ConcurrentGroup. Don't forget, Groups are also Tasks.
  g1 := task.NewConcurrentGroup()

  // Calling AddChild with a Task adds it to the group.
  // Here, we add 3 tasks. Each print their name.
  g1.AddChild(newPrintTask("task1"))
  g1.AddChild(newPrintTask("task2"))
  g1.AddChild(newPrintTask("task3"))

  // Run all 3 tasks concurrently.
  // Although each task takes about 1 second to run, it should all still
  // take about second because they are run concurrently (and spend most of its
  // time just sleeping).
  g1.Run(nil)
}

func newPrintTask(str string) task.Task {
  run := func(t task.Task, ctx task.Context) {
    time.Sleep(1 * time.Second)
    fmt.Println(str)
  }
  return task.NewTaskWithFunc(run)
}
```

You can also control, for a given ```ConcurrentGroup```, the maximum number of
tasks you want to be executing simultaneously at any given moment. You just
need to set the ```MaxConcurrency``` property of the ```ConcurrentGroup```
before running it.

```go
  g1 := task.NewConcurrentGroup()
  g1.MaxConcurrency = 100
```

For a more elaborate example please check
[examples/maxconcurrency.go](https://github.com/imkira/go-task/blob/master/examples/maxconcurrency.go).

#### Example: SerialGroup

If you want to run tasks in series, you can do it like:

```go
func main() {
  // Create a SerialGroup. Don't forget, Groups are also Tasks.
  g1 := task.NewSerialGroup()

  // Calling AddChild with a Task adds it to the group.
  // Here, we add 3 tasks. Each print their name.
  g1.AddChild(newPrintTask("task1"))
  g1.AddChild(newPrintTask("task2"))
  g1.AddChild(newPrintTask("task3"))

  // Run all 3 tasks in series.
  // Since each task takes about 1 second to execute, it should all take
  // about 3 seconds.
  g1.Run(nil)
}

func newPrintTask(str string) task.Task {
  run := func(t task.Task, ctx task.Context) {
    time.Sleep(1 * time.Second)
    fmt.Println(str)
  }
  return task.NewTaskWithFunc(run)
}
```

### Example: Task Trees

Because Groups are also Tasks, you can compose tasks into trees of tasks.
Please check
[examples/tree.go](https://github.com/imkira/go-task/blob/master/examples/tree.go)
for an example on how to do it.

### Example: Waiting for Tasks

Suppose you want to know when a particular task finished executing.

You can do it for individual Tasks or for task Groups.

```go
  // create t1, t2, t3 tasks
  // ...

  // here we create a ConcurrentGroup and we add 3 tasks: t1, t2, and t3.
  g1 := task.NewConcurrentGroup()
  g1.AddChild(t1)
  g1.AddChild(t2)
  g1.AddChild(t3)

  // run the group in a separate goroutine so it doesn't block here
  go g1.Run(nil)

  // wait for Task t2 to finish
  err1 := t2.Wait()
  // err1 will be returned if Task t2 is cancelled.

  // wait for the whole Group g1 to finish
  err2 := g1.Wait()
  // err2 will be returned if Task t2 is cancelled.
```

### Example: Cancelling Tasks

Sometimes you want to abort execution of a task. In case of a group, if a task
is cancelled, the group gets cancelled too as soon as possible.

#### Example: Cancelling Tasks Explicitely

```go
  // create longTask
  // ...

  // run a long task
  go longTask.Run(nil)

  // cancel task 1 second later
  time.AfterFunc(time.Second, func() {
    // Cancel task with custom error.
    // You can also pass nil (it will become task.ErrTaskCancelled).
    longTask.Cancel(fooError)
  })

  // wait for longTask to finish executing
  err := longTask.Wait()
  // err will be the same as fooError if the task was cancelled
```

#### Example: Cancelling Tasks with Timeouts

```go
  // create longTask
  // ...

  // set a maximum duration of 3 seconds for this task.
  longTask.SetTimeout(3 * time.Second)

  // run and wait for a long task
  longTask.Run(nil)

  err := longTask.Err()
  // err will be the ErrTaskDeadlineExceeded if the task took more than 3
  // seconds to execute.
```

#### Example: Cancelling Tasks with Deadlines

```go
  // create longTask
  // ...

  // set a maximum duration of 3 seconds for this task.
  task2.SetDeadline(time.Now().Add(3 * time.Second))

  // run and wait for a long task
  longTask.Run(nil)

  err := longTask.Err()
  // err will be the ErrTaskDeadlineExceeded if the task took more than 3
  // seconds to execute.
```

Please note this example is "almost" equivalent to the previous. But there is
actually a big difference: With SetTimeout, the deadline is automatically set
right when the task is about to run. With SetDeadline, the deadline is set the
moment you call SetDeadline. For this reason, you should use SetTimeout for
task duration restrictions (e.g., the task should not take more than X time),
and SetDeadline for absolute date/time restrictions (e.g., the task should
finish before midnight).

#### Example: Cancelling Tasks with Signals

```go
  // create longTask
  // ...

  // cancel task if SIGTERM and SIGINT (ctrl-c) is detected
  task.SetSignals(syscall.SIGINT, syscall.SIGTERM)

  // run and wait for a long task
  longTask.Run(nil)

  err := longTask.Err()
  // err will be the ErrTaskSignalReceived if a signal is received while the
  // task is running.
```

### Example: Sharing Data with Contexts

[Contexts](http://godoc.org/github.com/imkira/go-task#Context) are useful for
sharing data with tasks.

```go
func main() {
  // create a Context object
  ctx := task.NewContext()

  // put what you want inside
  ctx.Set("foo", "bar")
  ctx.Set("true", true)
  ctx.Set("one", 1)
  ctx.Set("pi", 3.14)

  // create task that prints the context
  t := newPrintContextTask()

  // run it using the context
  t.Run(ctx)
}

func newPrintContextTask() task.Task {
  run := func(t task.Task, ctx task.Context) {
    fmt.Println(ctx.Get("foo"))
    fmt.Println(ctx.Get("true"))
    fmt.Println(ctx.Get("one"))
    fmt.Println(ctx.Get("pi"))
  }
  return task.NewTaskWithFunc(run)
}
```

### More Examples

For more elaborate examples please check
[examples](https://github.com/imkira/go-task/blob/master/examples).

## Contribute

Found a bug? Want to contribute and add a new feature?

Please fork this project and send me a pull request!

## License

go-task is licensed under the MIT license:

www.opensource.org/licenses/MIT

## Copyright

Copyright (c) 2015 Mario Freitas. See
[LICENSE.txt](http://github.com/imkira/go-task/blob/master/LICENSE.txt)
for further details.
