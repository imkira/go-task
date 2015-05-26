// Execute 1000 tasks concurrently but limit maximum concurrency to 100
// simultaneous goroutines.
package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/imkira/go-task"
)

var numTasks, curTasks int64

func newPrintTask(str string) task.Task {
	run := func(t task.Task, ctx task.Context) {
		atomic.AddInt64(&numTasks, 1)
		atomic.AddInt64(&curTasks, 1)
		defer atomic.AddInt64(&curTasks, -1)

		fmt.Printf("%s (total: %d, now: %d)\n", str, numTasks, curTasks)

		// sleep between 500 and 1000ms
		duration := time.Duration(500+(int)(rand.Float64()*500)) * time.Millisecond
		time.Sleep(duration)
	}
	return task.NewTaskWithFunc(run)
}

func main() {
	group := task.NewConcurrentGroup()
	group.MaxConcurrency = 100

	for i := 1; i <= 1000; i++ {
		group.AddChild(newPrintTask(fmt.Sprintf("task: %d", i)))
	}

	group.Run(nil)
}
