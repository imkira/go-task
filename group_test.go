package task

import (
	"errors"
	"testing"
	"time"
)

func TestNewConcurrentGroup(t *testing.T) {
	g := NewConcurrentGroup()
	if g == nil {
		t.Fatalf("Not expecting nil")
	}
	if len(g.children) != 0 {
		t.Fatalf("Invalid children")
	}
	if g.MaxConcurrency != 0 {
		t.Fatalf("Invalid context")
	}
}

func TestConcurrentGroupAddChild(t *testing.T) {
	t1 := newTestTask()
	t2 := newTestTask()
	t3 := newTestTask()
	g := NewConcurrentGroup()
	g.AddChild(t1.task)
	if len(g.children) != 1 {
		t.Fatalf("Invalid tasks")
	}
	if g.children[0] == nil {
		t.Fatalf("Invalid tasks")
	}
	g.AddChild(t2.task)
	if len(g.children) != 2 {
		t.Fatalf("Invalid tasks")
	}
	if g.children[0] == nil || g.children[1] == nil {
		t.Fatalf("Invalid tasks")
	}
	g.AddChild(t3.task)
	if len(g.children) != 3 {
		t.Fatalf("Invalid tasks")
	}
	if g.children[0] == nil || g.children[1] == nil || g.children[2] == nil {
		t.Fatalf("Invalid tasks")
	}
}

func TestConcurrentGroupRunMultipleReturnsError(t *testing.T) {
	t1 := newTestTask()
	t2 := newTestTask()
	t3 := newTestTask()
	g := NewConcurrentGroup()
	g.AddChild(t1.task)
	g.AddChild(t2.task)
	g.AddChild(t3.task)

	testTaskRunMultipleReturnsError(t, g)

	if len(t1.entries) != 1 || len(t2.entries) != 1 || len(t3.entries) != 1 {
		t.Fatalf("Invalid call count")
	}
	if len(t1.returns) != 1 || len(t2.returns) != 1 || len(t3.returns) != 1 {
		t.Fatalf("Invalid call count")
	}
}

func TestConcurrentGroupRunRunsConcurrently(t *testing.T) {
	g := NewConcurrentGroup()
	t1 := newTestTask()
	t1.sleep = 500 * time.Millisecond
	t2 := newTestTask()
	t2.sleep = 800 * time.Millisecond
	t3 := newTestTask()
	t3.sleep = 500 * time.Millisecond
	g.AddChild(t1.task)
	g.AddChild(t2.task)
	g.AddChild(t3.task)
	start := time.Now()
	g.Run(nil)
	end := time.Now()
	if err := g.Err(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	minimum := testMaxDuration(t1.sleep, testMaxDuration(t2.sleep, t3.sleep))
	threshold := 200 * time.Millisecond
	if end.Sub(start) < minimum {
		t.Fatalf("Too fast")
	}
	if end.Sub(start) > (minimum + threshold) {
		t.Fatalf("Too slow")
	}
	if len(t1.returns) != 1 || len(t2.returns) != 1 || len(t3.returns) != 1 {
		t.Fatalf("Invalid call count")
	}
	if testDuration(t1.entries[0], t2.entries[0]) > threshold {
		t.Fatalf("Invalid entry times")
	}
	if testDuration(t2.entries[0], t3.entries[0]) > threshold {
		t.Fatalf("Invalid entry times")
	}
}

func TestConcurrentGroupRunReturnsFirstError(t *testing.T) {
	g := NewConcurrentGroup()
	t1 := newTestTask()
	t1.sleep = 500 * time.Millisecond
	t1.err = errors.New("test error")
	t2 := newTestTask()
	t2.err = errors.New("test error")
	t2.sleep = 100 * time.Millisecond
	t3 := newTestTask()
	t3.err = errors.New("test error")
	t3.sleep = 700 * time.Millisecond
	g.AddChild(t1.task)
	g.AddChild(t2.task)
	g.AddChild(t3.task)
	g.Run(nil)
	if err := g.Err(); err != t2.err {
		t.Fatalf("Invalid error: %v", err)
	}
	if len(t1.returns) != 1 || len(t2.returns) != 1 || len(t3.returns) != 1 {
		t.Fatalf("Invalid call count")
	}
}

func TestConcurrentGroupRunLimitsMaxConcurrency1(t *testing.T) {
	g := NewConcurrentGroup()
	t1 := newTestTask()
	t1.sleep = 500 * time.Millisecond
	t2 := newTestTask()
	t2.sleep = 800 * time.Millisecond
	t3 := newTestTask()
	t3.sleep = 500 * time.Millisecond
	g.AddChild(t1.task)
	g.AddChild(t2.task)
	g.AddChild(t3.task)
	g.MaxConcurrency = 1
	start := time.Now()
	g.Run(nil)
	end := time.Now()
	if err := g.Err(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	minimum := t1.sleep + t2.sleep + t3.sleep
	threshold := 200 * time.Millisecond
	if end.Sub(start) < minimum {
		t.Fatalf("Too fast")
	}
	if end.Sub(start) > (minimum + threshold) {
		t.Fatalf("Too slow")
	}
	if len(t1.returns) != 1 || len(t2.returns) != 1 || len(t3.returns) != 1 {
		t.Fatalf("Invalid call count")
	}
	if t1.returns[0].After(t2.entries[0]) {
		t.Fatalf("MaxConcurrency failure")
	}
	if t2.returns[0].After(t3.entries[0]) {
		t.Fatalf("MaxConcurrency failure")
	}
}

func TestConcurrentGroupRunLimitsMaxConcurrency2(t *testing.T) {
	g := NewConcurrentGroup()
	t1 := newTestTask()
	t1.sleep = 500 * time.Millisecond
	t2 := newTestTask()
	t2.sleep = 800 * time.Millisecond
	t3 := newTestTask()
	t3.sleep = 500 * time.Millisecond
	g.AddChild(t1.task)
	g.AddChild(t2.task)
	g.AddChild(t3.task)
	g.MaxConcurrency = 2
	start := time.Now()
	g.Run(nil)
	end := time.Now()
	if err := g.Err(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	minimum := t1.sleep + t3.sleep
	threshold := 200 * time.Millisecond
	if end.Sub(start) < minimum {
		t.Fatalf("Too fast")
	}
	if end.Sub(start) > (minimum + threshold) {
		t.Fatalf("Too slow")
	}
	if len(t1.returns) != 1 || len(t2.returns) != 1 || len(t3.returns) != 1 {
		t.Fatalf("Invalid call count")
	}
	if t1.returns[0].After(t3.entries[0]) {
		t.Fatalf("MaxConcurrency failure")
	}
	if t3.entries[0].After(t2.returns[0]) {
		t.Fatalf("MaxConcurrency failure")
	}
}

func TestConcurrentGroupRunPassesContext(t *testing.T) {
	dchan := make(chan Context, 3)
	run := func(t Task, d Context) {
		dchan <- d
	}
	g := NewConcurrentGroup()
	for i := 0; i < cap(dchan); i++ {
		g.AddChild(NewTaskWithFunc(run))
	}
	d := NewContext()
	g.Run(d)
	if err := g.Err(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	for i := 0; i < cap(dchan); i++ {
		if d2 := <-dchan; d2 != d {
			t.Fatalf("Invalid ctx")
		}
	}
}

func TestNewSerialGroup(t *testing.T) {
	g := NewSerialGroup()
	if g == nil {
		t.Fatalf("Not expecting nil")
	}
	if len(g.children) != 0 {
		t.Fatalf("Invalid tasks")
	}
}

func TestSerialGroupAddChild(t *testing.T) {
	t1 := newTestTask()
	t2 := newTestTask()
	t3 := newTestTask()
	g := NewSerialGroup()
	g.AddChild(t1.task)
	if len(g.children) != 1 {
		t.Fatalf("Invalid tasks")
	}
	if g.children[0] == nil {
		t.Fatalf("Invalid tasks")
	}
	g.AddChild(t2.task)
	if len(g.children) != 2 {
		t.Fatalf("Invalid tasks")
	}
	if g.children[0] == nil || g.children[1] == nil {
		t.Fatalf("Invalid tasks")
	}
	g.AddChild(t3.task)
	if len(g.children) != 3 {
		t.Fatalf("Invalid tasks")
	}
	if g.children[0] == nil || g.children[1] == nil || g.children[2] == nil {
		t.Fatalf("Invalid tasks")
	}
}

func TestSerialGroupRunMultipleReturnsError(t *testing.T) {
	t1 := newTestTask()
	t2 := newTestTask()
	t3 := newTestTask()
	g := NewSerialGroup()
	g.AddChild(t1.task)
	g.AddChild(t2.task)
	g.AddChild(t3.task)

	testTaskRunMultipleReturnsError(t, g)

	if len(t1.entries) != 1 || len(t2.entries) != 1 || len(t3.entries) != 1 {
		t.Fatalf("Invalid call count")
	}
	if len(t1.returns) != 1 || len(t2.returns) != 1 || len(t3.returns) != 1 {
		t.Fatalf("Invalid call count")
	}
}

func TestSerialGroupRunRunsSerially(t *testing.T) {
	g := NewSerialGroup()
	t1 := newTestTask()
	t1.sleep = 500 * time.Millisecond
	t2 := newTestTask()
	t2.sleep = 800 * time.Millisecond
	t3 := newTestTask()
	t3.sleep = 500 * time.Millisecond
	g.AddChild(t1.task)
	g.AddChild(t2.task)
	g.AddChild(t3.task)
	start := time.Now()
	g.Run(nil)
	end := time.Now()
	if err := g.Err(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	minimum := t1.sleep + t2.sleep + t3.sleep
	threshold := 200 * time.Millisecond
	if end.Sub(start) < minimum {
		t.Fatalf("Too fast")
	}
	if end.Sub(start) > (minimum + threshold) {
		t.Fatalf("Too slow")
	}
	if len(t1.returns) != 1 || len(t2.returns) != 1 || len(t3.returns) != 1 {
		t.Fatalf("Invalid call count")
	}
	if t2.entries[0].Before(t1.returns[0]) {
		t.Fatalf("Invalid entry times")
	}
	if t3.entries[0].Before(t1.returns[0]) {
		t.Fatalf("Invalid entry times")
	}
	if t3.entries[0].Before(t2.returns[0]) {
		t.Fatalf("Invalid entry times")
	}
}

func TestSerialGroupRunReturnsFirstError(t *testing.T) {
	g := NewSerialGroup()
	t1 := newTestTask()
	t1.sleep = 500 * time.Millisecond
	t1.err = errors.New("test error")
	t2 := newTestTask()
	t2.err = errors.New("test error")
	t2.sleep = 100 * time.Millisecond
	t3 := newTestTask()
	t3.err = errors.New("test error")
	t3.sleep = 700 * time.Millisecond
	g.AddChild(t1.task)
	g.AddChild(t2.task)
	g.AddChild(t3.task)
	g.Run(nil)
	if err := g.Err(); err != t1.err {
		t.Fatalf("Invalid error: %v", err)
	}
	if len(t1.returns) != 1 || len(t2.returns) != 0 || len(t3.returns) != 0 {
		t.Fatalf("Invalid call count")
	}
}

func TestSerialGroupRunPassesContext(t *testing.T) {
	dchan := make(chan Context, 3)
	run := func(t Task, d Context) {
		dchan <- d
	}
	g := NewSerialGroup()
	for i := 0; i < cap(dchan); i++ {
		g.AddChild(NewTaskWithFunc(run))
	}
	d := NewContext()
	g.Run(d)
	if err := g.Err(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	for i := 0; i < cap(dchan); i++ {
		if d2 := <-dchan; d2 != d {
			t.Fatalf("Invalid ctx")
		}
	}
}

func testMaxDuration(d1, d2 time.Duration) time.Duration {
	if d1 > d2 {
		return d1
	}
	return d2
}

func testDuration(t1, t2 time.Time) time.Duration {
	if t1.After(t2) {
		return t1.Sub(t2)
	}
	return t2.Sub(t1)
}
