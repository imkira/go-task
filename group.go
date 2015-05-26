package task

import "sync"

// ConcurrentGroup is a task that runs its sub-tasks concurrently.
// ConcurrentGroup implements the Task interface so you can compose it with
// other tasks.
type ConcurrentGroup struct {
	*StandardTask

	// MaxConcurrency is a limit to the maximum number of goroutines that can
	// execute simultaneously. If you set it above 0 this limit is enforced,
	// otherwise it is unlimited. By default, it is set to 0 (unlimited).
	MaxConcurrency int
}

// NewConcurrentGroup creates a new ConcurrentGroup.
func NewConcurrentGroup() *ConcurrentGroup {
	return &ConcurrentGroup{
		StandardTask: NewStandardTask(),
	}
}

// Run runs all tasks concurrently, according to the maximum concurrency
// setting, and waits until all of them are done. On the first error returned
// by one of the child tasks, the group will be cancelled. If, in any way, the
// group is cancelled, no more child tasks will be run. This function will
// always wait until no more child tasks in the group are running.
func (g *ConcurrentGroup) Run(ctx Context) {
	if err := g.Start(); err != nil {
		return
	}
	defer g.Finish()

	// run each task in a separate goroutine
	sem := newSemaphore(g.MaxConcurrency)
	wg := &sync.WaitGroup{}
	wg.Add(len(g.children))
	for _, child := range g.children {
		sem.Lock()
		go g.run(child, ctx, sem, wg)
	}

	// wait for all tasks
	wg.Wait()
}

func (g *ConcurrentGroup) run(task Task, ctx Context, sem semaphore, wg *sync.WaitGroup) {
	defer wg.Done()
	defer sem.Unlock()
	task.Run(ctx)
	err := task.Err()
	if err != nil {
		g.Cancel(err)
	}
}

// SerialGroup is a task that runs its sub-tasks serially (in sequence).
// SerialGroup implements the Task interface so you can compose it with other
// tasks.
type SerialGroup struct {
	*StandardTask
}

// NewSerialGroup creates a new SerialGroup.
func NewSerialGroup() *SerialGroup {
	return &SerialGroup{
		StandardTask: NewStandardTask(),
	}
}

// Run runs all tasks serially and waits until all of them are done. On the
// first error returned by one of the tasks, no more child tasks are run, and
// the group is cancelled allowing currently executing tasks to cleanup and
// finish.
func (g *SerialGroup) Run(ctx Context) {
	if err := g.Start(); err != nil {
		return
	}
	defer g.Finish()

	for _, child := range g.children {
		child.Run(ctx)
		if err := child.Err(); err != nil {
			g.Cancel(err)
			return
		}
	}
}
