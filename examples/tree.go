// Creates the following task tree:
//
// task1
// task2
//   task3
//   task4
//   task5
//   task6
//     task7
//     task8
// task9
// task10
package main

import (
	"fmt"
	"time"

	"github.com/imkira/go-task"
)

func newPrintTask(str string) task.Task {
	run := func(t task.Task, ctx task.Context) {
		time.Sleep(2 * time.Second)
		fmt.Println(str)
	}
	return task.NewTaskWithFunc(run)
}

func main() {
	g1 := task.NewSerialGroup()
	g1.AddChild(newPrintTask("|-task1"))
	g1.AddChild(newPrintTask("|-task2"))

	g2 := task.NewConcurrentGroup()
	t3 := newPrintTask("|--task3")
	g2.AddChild(t3)
	g2.AddChild(newPrintTask("|--task4"))
	g2.AddChild(newPrintTask("|--task5"))
	g2.AddChild(newPrintTask("|--task6"))
	g1.AddChild(g2)

	g3 := task.NewSerialGroup()
	g3.AddChild(newPrintTask("|---task7"))
	g3.AddChild(newPrintTask("|---task8"))
	g2.AddChild(g3)

	g4 := task.NewConcurrentGroup()
	g4.AddChild(newPrintTask("|-task9"))
	g4.AddChild(newPrintTask("|-task10"))
	g1.AddChild(g4)

	go g1.Run(nil)

	// wait for individual task to finish
	err := t3.Wait()
	fmt.Printf("task3 wait: %v\n", err)

	// wait for group task to finish
	err = g2.Wait()
	fmt.Printf("group2 wait: %v\n", err)

	// wait for top task
	err = g1.Wait()
	fmt.Printf("group1 wait: %v\n", err)
}
