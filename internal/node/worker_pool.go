package node

// WorkerPool is a placeholder for a pool of workers running simulation tasks.
// TODO: implement proper worker lifecycle and job dispatch.
type WorkerPool struct {
	size int
}

// NewWorkerPool creates a new WorkerPool with the given size.
func NewWorkerPool(size int) *WorkerPool {
	return &WorkerPool{size: size}
}

// Start starts the worker pool.
// This is currently a no-op placeholder.
func (p *WorkerPool) Start() {
	// TODO: start worker goroutines here
}

// Stop stops the worker pool.
// This is currently a no-op placeholder.
func (p *WorkerPool) Stop() {
	// TODO: signal workers to stop and wait for them
}
