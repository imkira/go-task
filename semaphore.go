package task

var (
	defaultNilSemaphore = &nilSemaphore{}
	semaphoreSignal     = struct{}{}
)

type semaphore interface {
	Lock()
	Unlock()
}

func newSemaphore(limit int) semaphore {
	if limit <= 0 {
		return defaultNilSemaphore
	}
	return newStandardSemaphore(limit)
}

type nilSemaphore struct {
}

func (s *nilSemaphore) Lock() {
}

func (s *nilSemaphore) Unlock() {
}

var _ semaphore = (*nilSemaphore)(nil)

type standardSemaphore struct {
	signalChan chan struct{}
}

func newStandardSemaphore(limit int) *standardSemaphore {
	signalChan := make(chan struct{}, limit)
	for i := 0; i < limit; i++ {
		signalChan <- semaphoreSignal
	}
	return &standardSemaphore{
		signalChan: signalChan,
	}
}

func (s *standardSemaphore) Lock() {
	<-s.signalChan
}

func (s *standardSemaphore) Unlock() {
	s.signalChan <- semaphoreSignal
}

var _ semaphore = (*standardSemaphore)(nil)
