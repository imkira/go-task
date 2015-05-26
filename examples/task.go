// Creates tasks t1, t2, t3 and cancels t2. Waits for them to finish and prints
// state information about each.
package main

import (
	"fmt"
	"time"

	"github.com/imkira/go-task"
)

func main() {
	// create task1
	t1 := newPrintTask("task1")

	// create task2
	t2 := newPrintTask("task2")

	// create task3
	t3 := newPrintTask("task3")

	// create a group that runs t1, t2 and t3 in sequence
	g := task.NewSerialGroup()
	g.AddChild(t1)
	g.AddChild(t2)
	g.AddChild(t3)

	// cancel t3 one second right after running the group
	time.AfterFunc(time.Second, func() {
		t3.Cancel(nil)
	})

	// run the group and wait for it
	g.Run(nil)
	fmt.Printf("g err: %v\n", g.Err())

	// show status for task1
	showStatus("task1", t1)

	// show status for task2
	showStatus("task2", t2)

	// show status for task3
	showStatus("task3", t2)
}

func newPrintTask(str string) task.Task {
	run := func(t task.Task, ctx task.Context) {
		fmt.Printf("%s: starting...\n", str)
		time.Sleep(2 * time.Second)
		fmt.Printf("%s: finishing...\n", str)
	}
	return task.NewTaskWithFunc(run)
}

func showStatus(name string, t task.Task) {
	select {
	case <-t.Started():
		fmt.Printf("%s did start\n", name)
	default:
		fmt.Printf("%s did not start\n", name)
	}

	select {
	case <-t.Cancelled():
		fmt.Printf("%s was cancelled\n", name)
	default:
		fmt.Printf("%s was not cancelled\n", name)
	}

	select {
	case <-t.Finished():
		fmt.Printf("%s did finish\n", name)
	default:
		fmt.Printf("%s did not finish\n", name)
	}
}
