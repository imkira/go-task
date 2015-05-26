package task

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"
)

type testTask struct {
	task Task

	sync.Mutex
	entries []time.Time
	returns []time.Time
	sleep   time.Duration
	err     error
}

func newTestTask() *testTask {
	t := &testTask{}
	run := func(t2 Task, ctx Context) {
		t.Lock()
		t.entries = append(t.entries, time.Now())
		t.Unlock()
		if t.sleep > 0 {
			time.Sleep(t.sleep)
		}
		t.Lock()
		t.returns = append(t.returns, time.Now())
		t.Unlock()

		if t.err != nil {
			t2.Cancel(t.err)
		}
	}

	t.task = NewTaskWithFunc(run)
	return t
}

func TestNewStandardTask(t *testing.T) {
	task := NewStandardTask()
	if task == nil {
		t.Fatalf("Not expecting nil")
	}

	if len(task.children) != 0 {
		t.Fatalf("Invalid children")
	}

	select {
	case <-task.Started():
		t.Fatalf("Invalid task")
	default:
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	if task.Err() != nil {
		t.Fatalf("Unexpected error")
	}
}

func TestStandardTaskStartMultipleReturnsError(t *testing.T) {
	task := NewStandardTask()

	if err := task.Start(); err != nil {
		t.Fatalf("Unexpected error")
	}

	for i := 0; i < 2; i++ {
		if err := task.Start(); err != ErrInvalidTaskState {
			t.Fatalf("Expecting error due to invalid state")
		}

		select {
		case <-task.Started():
		default:
			t.Fatalf("Invalid task")
		}

		select {
		case <-task.Cancelled():
			t.Fatalf("Invalid task")
		default:
		}

		if task.Err() != nil {
			t.Fatalf("Unexpected error")
		}

		select {
		case <-task.Finished():
			t.Fatalf("Invalid task")
		default:
		}
	}

	if err2 := task.Finish(); err2 != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	if task.Err() != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Finished():
	default:
		t.Fatalf("Invalid task")
	}
}

func TestStandardTaskCancelDueToTimeout(t *testing.T) {
	task := NewStandardTask()

	task.SetTimeout(500 * time.Millisecond)

	// time is not used until task is actually started
	time.Sleep(1000 * time.Millisecond)

	if err := task.Start(); err != nil {
		t.Fatalf("Unexpected error")
	}

	// a little before timing out
	time.Sleep(400 * time.Millisecond)

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	if task.Err() != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	// must have timed out
	time.Sleep(200 * time.Millisecond)

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
	default:
		t.Fatalf("Invalid task")
	}

	if task.Err() != ErrTaskDeadlineExceeded {
		t.Fatalf("Invalid error")
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	if err2 := task.Finish(); err2 != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
	default:
		t.Fatalf("Invalid task")
	}

	if task.Err() != ErrTaskDeadlineExceeded {
		t.Fatalf("Invalid error")
	}

	select {
	case <-task.Finished():
	default:
		t.Fatalf("Invalid task")
	}
}

func TestStandardTaskCancelDueToSignal(t *testing.T) {
	task := NewStandardTask()

	task.SetSignals(syscall.SIGUSR1)

	if err := task.Start(); err != nil {
		t.Fatalf("Unexpected error")
	}

	// wait a bit (still OK)
	time.Sleep(400 * time.Millisecond)

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	if task.Err() != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	// send signal
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)

	// give it a little time ot handle the signal notification
	time.Sleep(200 * time.Millisecond)

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
	default:
		t.Fatalf("Invalid task")
	}

	if task.Err() != ErrTaskSignalReceived {
		t.Fatalf("Invalid error")
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	if err2 := task.Finish(); err2 != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
	default:
		t.Fatalf("Invalid task")
	}

	if task.Err() != ErrTaskSignalReceived {
		t.Fatalf("Invalid error")
	}

	select {
	case <-task.Finished():
	default:
		t.Fatalf("Invalid task")
	}
}

func TestStandardTaskCancelNilSetsDefaultError(t *testing.T) {
	task := NewStandardTask()
	task.Cancel(nil)
	if task.Err() != ErrTaskCancelled {
		t.Fatalf("Invalid error")
	}
}

func TestStandardTaskCreatedCancelled(t *testing.T) {
	task := NewStandardTask()
	err := errors.New("test error")

	go func() {
		time.Sleep(1000 * time.Millisecond)
		task.Cancel(err)
	}()

	task.Wait()

	if task.Err() != err {
		t.Fatalf("Invalid error")
	}

	for i := 0; i < 2; i++ {
		select {
		case <-task.Started():
			t.Fatalf("Invalid task")
		default:
		}

		select {
		case <-task.Cancelled():
		default:
			t.Fatalf("Invalid task")
		}

		if task.Err() != err {
			t.Fatalf("Invalid error")
		}

		select {
		case <-task.Finished():
			t.Fatalf("Invalid task")
		default:
		}

		if err2 := task.Finish(); err2 != ErrInvalidTaskState {
			t.Fatalf("Invalid error")
		}
	}
}

func testStandardTaskStartingCancelled(t *testing.T, task *StandardTask) {
	if err := task.Start(); err != ErrTaskDeadlineExceeded {
		t.Fatalf("Expecting error due to deadline")
	}

	task.Wait()

	for i := 0; i < 2; i++ {
		select {
		case <-task.Started():
			t.Fatalf("Invalid task")
		default:
		}

		select {
		case <-task.Cancelled():
		default:
			t.Fatalf("Invalid task")
		}

		if task.Err() != ErrTaskDeadlineExceeded {
			t.Fatalf("Invalid error")
		}

		select {
		case <-task.Finished():
			t.Fatalf("Invalid task")
		default:
		}

		if err2 := task.Finish(); err2 != ErrInvalidTaskState {
			t.Fatalf("Invalid error")
		}
	}

	if task.Err() != ErrTaskDeadlineExceeded {
		t.Fatalf("Invalid error")
	}
}

func TestStandardTaskStartingCancelled(t *testing.T) {
	task1 := NewStandardTask()
	task1.SetTimeout(0)
	testStandardTaskStartingCancelled(t, task1)

	task2 := NewStandardTask()
	task2.SetDeadline(time.Now())
	testStandardTaskStartingCancelled(t, task2)
}

func TestStandardTaskStartedCancelled(t *testing.T) {
	task := NewStandardTask()
	task.SetSignals(syscall.SIGUSR1)

	if err := task.Start(); err != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	err := errors.New("test error")
	task.Cancel(err)

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
	default:
		t.Fatalf("Invalid task")
	}

	if task.Err() != err {
		t.Fatalf("Invalid error")
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	errChan := make(chan error, 1)
	go func() {
		time.Sleep(1000 * time.Millisecond)
		errChan <- task.Finish()
	}()

	task.Wait()

	if err2 := <-errChan; err2 != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
	default:
		t.Fatalf("Invalid task")
	}

	if task.Err() != err {
		t.Fatalf("Invalid error")
	}

	select {
	case <-task.Finished():
	default:
		t.Fatalf("Invalid task")
	}
}

func TestStandardTaskFinishedCancel(t *testing.T) {
	task := NewStandardTask()
	if err := task.Start(); err != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	if err2 := task.Finish(); err2 != nil {
		t.Fatalf("Unexpected error")
	}

	task.Wait()

	err := errors.New("test error")
	task.Cancel(err)

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	if task.Err() != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Finished():
	default:
		t.Fatalf("Invalid task")
	}
}

func TestStandardTaskFinishedSignal(t *testing.T) {
	task := NewStandardTask()
	task.SetSignals(syscall.SIGUSR1)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)
	defer signal.Stop(ch)

	if err := task.Start(); err != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	if err2 := task.Finish(); err2 != nil {
		t.Fatalf("Unexpected error")
	}

	// send signal
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)

	// give it a little time ot handle the signal notification
	time.Sleep(200 * time.Millisecond)

	task.Wait()

	err := errors.New("test error")
	task.Cancel(err)

	select {
	case <-task.Started():
	default:
		t.Fatalf("Invalid task")
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	if task.Err() != nil {
		t.Fatalf("Unexpected error")
	}

	select {
	case <-task.Finished():
	default:
		t.Fatalf("Invalid task")
	}
}

func TestFuncTaskRunRunsTask(t *testing.T) {
	t1 := newTestTask()
	task := t1.task

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	if task.Err() != nil {
		t.Fatalf("Invalid error")
	}

	select {
	case <-task.Finished():
		t.Fatalf("Invalid task")
	default:
	}

	task.Run(nil)

	if len(t1.entries) != 1 || len(t1.returns) != 1 {
		t.Fatalf("Invalid call count")
	}

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	if task.Err() != nil {
		t.Fatalf("Invalid error")
	}

	select {
	case <-task.Finished():
	default:
		t.Fatalf("Invalid task")
	}
}

func TestFuncTaskRunPassesTask(t *testing.T) {
	tchan := make(chan Task, 1)
	run := func(t Task, d Context) {
		tchan <- t
	}
	task := NewTaskWithFunc(run)
	task.Run(nil)
	if t2 := <-tchan; t2 != task {
		t.Fatalf("Invalid task")
	}
}

func TestFuncTaskRunPassesContext(t *testing.T) {
	dchan := make(chan Context, 1)
	run := func(t Task, d Context) {
		dchan <- d
	}
	t1 := NewTaskWithFunc(run)
	t1.Run(nil)
	if d2 := <-dchan; d2 != nil {
		t.Fatalf("Invalid ctx")
	}
	d := NewContext()
	t2 := NewTaskWithFunc(run)
	t2.Run(d)
	if d2 := <-dchan; d2 != d {
		t.Fatalf("Invalid ctx")
	}
}

func testTaskRunMultipleReturnsError(t *testing.T, task Task) {
	task.Run(nil)
	task.Run(nil)

	select {
	case <-task.Cancelled():
		t.Fatalf("Invalid task")
	default:
	}

	if task.Err() != nil {
		t.Fatalf("Invalid error")
	}

	select {
	case <-task.Finished():
	default:
		t.Fatalf("Invalid task")
	}
}

func TestFuncTaskRunMultipleReturnsError(t *testing.T) {
	t1 := newTestTask()
	task := t1.task
	testTaskRunMultipleReturnsError(t, task)
	if len(t1.entries) != 1 || len(t1.returns) != 1 {
		t.Fatalf("Invalid call count")
	}
}
