package workerpool

import "sync"

type WorkerPool struct {
	workerCount int
	jobQueue    chan func()
	done        chan struct{}
	mu          sync.Mutex
}

func NewWorkerPool(workerCount, jobQueueSize int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan func(), jobQueueSize),
		done:        make(chan struct{}),
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		go wp.worker()
	}
}

func (wp *WorkerPool) worker() {
	for {
		select {
		case job := <-wp.jobQueue:
			job()
		case <-wp.done:
			return
		}
	}
}

func (wp *WorkerPool) Submit(job func()) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.jobQueue <- job
}

func (wp *WorkerPool) Stop() {
	close(wp.jobQueue)
	close(wp.done)
}
