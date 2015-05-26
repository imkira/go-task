package main

import (
	"fmt"

	"github.com/imkira/go-task"
)

func incrementTotalTask() task.Task {
	run := func(t task.Task, ctx task.Context) {
		total := ctx.Get("total").(int)
		ctx.Set("total", total+1)
	}
	return task.NewTaskWithFunc(run)
}

func main() {
	g := task.NewSerialGroup()
	g.AddChild(incrementTotalTask())
	g.AddChild(incrementTotalTask())
	g.AddChild(incrementTotalTask())

	ctx := task.NewContext()
	ctx.Set("total", 0)
	g.Run(ctx)

	fmt.Printf("total is %v\n", ctx.Get("total"))
}
