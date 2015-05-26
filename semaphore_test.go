package task

import (
	"sync"
	"testing"
	"time"
)

func TestNewSemaphore(t *testing.T) {
	sem, ok := newSemaphore(3).(*standardSemaphore)
	if sem == nil || !ok {
		t.Fatalf("Not expecting nil")
	}
	if cap(sem.signalChan) != 3 {
		t.Fatalf("Bad capacity")
	}
	if len(sem.signalChan) != 3 {
		t.Fatalf("Bad length")
	}
	sem2, ok := newSemaphore(0).(*nilSemaphore)
	if sem2 == nil || !ok {
		t.Fatalf("Not expecting nil")
	}
}

func TestStandardSemaphoreLockAll(t *testing.T) {
	sem := newSemaphore(3)
	wg := &sync.WaitGroup{}
	wg.Add(3)
	minimum := 800 * time.Millisecond
	start := time.Now()
	for i := 0; i < 3; i++ {
		go func() {
			time.Sleep(minimum)
			sem.Lock()
			wg.Done()
		}()
	}
	wg.Wait()
	end := time.Now()
	threshold := 200 * time.Millisecond
	if end.Sub(start) < minimum {
		t.Fatalf("Too fast")
	}
	if end.Sub(start) > (minimum + threshold) {
		t.Fatalf("Too slow")
	}
}

func TestStandardSemaphoreUnlockAll(t *testing.T) {
	sem := newSemaphore(3)
	for i := 0; i < 3; i++ {
		sem.Lock()
	}
	wg := &sync.WaitGroup{}
	wg.Add(3)
	minimum := 800 * time.Millisecond
	start := time.Now()
	for i := 0; i < 3; i++ {
		go func() {
			time.Sleep(minimum)
			sem.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
	end := time.Now()
	threshold := 200 * time.Millisecond
	if end.Sub(start) < minimum {
		t.Fatalf("Too fast")
	}
	if end.Sub(start) > (minimum + threshold) {
		t.Fatalf("Too slow")
	}
}

func TestStandardSemaphoreLockWaitsUnlock(t *testing.T) {
	sem := newSemaphore(3)
	start := time.Now()
	for i := 0; i < 3; i++ {
		sem.Lock()
	}
	end := time.Now()
	if end.Sub(start) > (200 * time.Millisecond) {
		t.Fatalf("Too slow")
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	minimum := 3 * 800 * time.Millisecond
	start = time.Now()
	go func() {
		for i := 0; i < 3; i++ {
			sem.Lock()
		}
		wg.Done()
	}()
	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(800 * time.Millisecond)
			sem.Unlock()
		}
		wg2.Done()
	}()
	wg.Wait()
	end = time.Now()
	wg2.Wait()
	threshold := 200 * time.Millisecond
	if end.Sub(start) < minimum {
		t.Fatalf("Too fast")
	}
	if end.Sub(start) > (minimum + threshold) {
		t.Fatalf("Too slow")
	}
}
