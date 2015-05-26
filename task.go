package task

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	// ErrTaskCancelled happens when the task can be cancelled (it is in a
	// cancellable state) and is cancelled.
	ErrTaskCancelled = errors.New("task cancelled")

	// ErrTaskDeadlineExceeded happens when a timeout or deadline is defined for
	// the task, and it is triggered before the task finishes executing.
	ErrTaskDeadlineExceeded = errors.New("task deadline exceeded")

	// ErrTaskSignalReceived happens when one or more signals defined in the task
	// are triggered by the OS, before the task finishes executing.
	ErrTaskSignalReceived = errors.New("task received signal")

	// ErrInvalidTaskState happens when the task tries to do to an invalid state
	// transition (eg., starting or finishing the same task twice, etc...).
	ErrInvalidTaskState = errors.New("invalid task state")
)

// Task is an interface that helps you define, run and control a task. You can
// set deadlines and timeouts, you can cancel it, or wait for the it until it
// finishes execution or is canceled before starting.
type Task interface {
	// AddChild adds a child task to this task.
	AddChild(child Task)

	// SetDeadline sets the time after which the task should be
	// automatically cancelled.
	SetDeadline(t time.Time)

	// SetTimeout sets the duration after which the task should be
	// automatically cancelled.
	SetTimeout(d time.Duration)

	// SetSignals sets the list of signals that should cancel this task should
	// they be triggered by the OS. Please note that an additional goroutine will
	// be created in order to detect the signals and cancel the task.
	SetSignals(s ...os.Signal)

	// Run runs the task.
	Run(ctx Context)

	// Start prepares the task before being used.
	Start() error

	// Started returns a channel that's closed when the task is started.
	Started() <-chan struct{}

	// Cancel cancels the task with an error. If nil is passed, the default
	// ErrTaskCancelled is used instead.
	Cancel(error)

	// Cancelled returns a channel that's closed when the task is cancelled.
	Cancelled() <-chan struct{}

	// Finish finishes the task after being used.
	Finish() error

	// Finished returns a channel that's closed when the task is finished.
	Finished() <-chan struct{}

	// Wait waits for the task to finish executing, and returns the resulting
	// error (if any). Wait also returns immediately if the task is cancelled
	// before even starting execution.
	Wait() error

	// Err returns the error (if any) the task was cancelled with.
	Err() error
}

type taskState int

const (
	taskCreated taskState = iota
	taskCreatedCancelled
	taskStarting
	taskStartingCancelled
	taskStarted
	taskStartedCancelled
	taskFinished
	taskCancelledFinished
)

// StandardTask is an interface that helps you more easily implement Tasks.
// StandardTask implements all functionality required by the Task interface,
// with the exception of the Run function which is not included. This makes it
// ideal for embedding StandardTask in your struct and then implementing the
// required Run function.
type StandardTask struct {
	children  []Task
	useTimer  bool
	deadline  time.Time
	timeout   time.Duration
	signals   []os.Signal
	started   chan struct{}
	cancelled chan struct{}
	finished  chan struct{}

	mutex *sync.RWMutex
	state taskState
	timer *time.Timer
	err   error
}

// NewStandardTask creates a new StandardTask object.
func NewStandardTask() *StandardTask {
	return &StandardTask{
		started:   make(chan struct{}),
		cancelled: make(chan struct{}),
		finished:  make(chan struct{}),
		mutex:     &sync.RWMutex{},
		state:     taskCreated,
	}
}

// AddChild adds a child task to this task.
func (st *StandardTask) AddChild(child Task) {
	st.children = append(st.children, child)
}

// SetDeadline sets the time after which the task should be automatically
// cancelled.
func (st *StandardTask) SetDeadline(t time.Time) {
	st.useTimer = true
	st.deadline = t
	st.timeout = 0
}

// SetTimeout sets the duration after which the task should be automatically
// cancelled.
func (st *StandardTask) SetTimeout(d time.Duration) {
	st.useTimer = true
	st.timeout = d
	st.deadline = time.Time{}
}

// SetSignals sets the list of signals that should cancel this task should they
// be triggered by the OS. Please note that an additional goroutine will be
// created in order to detect the signals and cancel the task.
func (st *StandardTask) SetSignals(s ...os.Signal) {
	st.signals = s
}

// Start prepares the task before being used.
func (st *StandardTask) Start() error {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	if st.state != taskCreated {
		return ErrInvalidTaskState
	}
	st.state = taskStarting
	if st.useTimer {
		timeout := st.timeout
		if !st.deadline.IsZero() {
			timeout = st.deadline.Sub(time.Now())
		}
		if timeout <= 0 {
			st.cancel(ErrTaskDeadlineExceeded)
			return ErrTaskDeadlineExceeded
		}
		st.timer = time.AfterFunc(timeout, st.deadlineCancel)
	}
	if len(st.signals) > 0 {
		signalled := make(chan os.Signal, 1)
		signal.Notify(signalled, st.signals...)
		go st.handleSignal(signalled)
	}
	st.state = taskStarted
	close(st.started)
	return nil
}

func (st *StandardTask) deadlineCancel() {
	st.Cancel(ErrTaskDeadlineExceeded)
}

func (st *StandardTask) handleSignal(ch chan os.Signal) {
	defer signal.Stop(ch)
	select {
	case <-ch:
		st.Cancel(ErrTaskSignalReceived)
	case <-st.cancelled:
	case <-st.finished:
	}
}

// Started returns a channel that's closed when the task is started.
func (st *StandardTask) Started() <-chan struct{} {
	return st.started
}

// Cancel cancels the task with an error. If nil is passed, the default
// ErrTaskCancelled is used instead.
func (st *StandardTask) Cancel(err error) {
	st.mutex.Lock()
	st.cancel(err)
	st.mutex.Unlock()
}

func (st *StandardTask) cancel(err error) {
	switch st.state {
	case taskCreated:
		st.state = taskCreatedCancelled
	case taskStarting:
		st.state = taskStartingCancelled
	case taskStarted:
		st.state = taskStartedCancelled
	default:
		return
	}
	if err == nil {
		err = ErrTaskCancelled
	}
	st.err = err
	for _, child := range st.children {
		child.Cancel(err)
	}
	st.cleanup()
	close(st.cancelled)
}

// Cancelled returns a channel that's closed when the task is cancelled.
func (st *StandardTask) Cancelled() <-chan struct{} {
	return st.cancelled
}

// Finish finishes the task after being used.
func (st *StandardTask) Finish() error {
	st.mutex.Lock()
	switch st.state {
	case taskStarted:
		st.state = taskFinished
	case taskStartedCancelled:
		st.state = taskCancelledFinished
	default:
		st.mutex.Unlock()
		return ErrInvalidTaskState
	}
	st.cleanup()
	close(st.finished)
	st.mutex.Unlock()
	return nil
}

func (st *StandardTask) cleanup() {
	if st.timer != nil {
		st.timer.Stop()
		st.timer = nil
	}
	st.children = nil
}

// Finished returns a channel that's closed when the task is finished.
func (st *StandardTask) Finished() <-chan struct{} {
	return st.finished
}

// Wait waits for the task to finish executing, and returns the resulting
// error (if any). Wait also returns immediately if the task is cancelled
// before even starting execution.
func (st *StandardTask) Wait() error {
	select {
	case <-st.cancelled:
		st.mutex.RLock()
		state := st.state
		st.mutex.RUnlock()
		if state >= taskStarted {
			<-st.finished
		}
	case <-st.finished:
	}
	return st.Err()
}

// Err returns the error (if any) the task was cancelled with.
func (st *StandardTask) Err() error {
	st.mutex.RLock()
	err := st.err
	st.mutex.RUnlock()
	return err
}

// Func is a function that implements the Run function of a task.
type Func func(task Task, ctx Context)

type funcTask struct {
	*StandardTask
	run Func
}

// NewTaskWithFunc takes a function and wraps it in a Task.
func NewTaskWithFunc(run Func) Task {
	return &funcTask{
		StandardTask: NewStandardTask(),
		run:          run,
	}
}

// Run wraps the task and executes the function associated to this task.
func (ft *funcTask) Run(ctx Context) {
	if err := ft.Start(); err != nil {
		return
	}
	defer ft.Finish()
	ft.run(ft, ctx)
}
