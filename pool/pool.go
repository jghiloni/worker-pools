package pool

import (
	"context"
	"sync"
)

// PrePostWorkFunc is a function that, if set, will be called at the beginning or end
// of a pool worker's lifecycle, and can be useful for setup or teardown of a given
// worker
type PrePostWorkFunc func(context.Context)

// WorkFunc is the type called by every pool worker when it receives an item to process
type WorkFunc[T any] func(context.Context, T)

// WorkerPool represents a constrained collection of goroutines that all pull from a
// shared channel and do the same thing, allowing for cleaner code around parallelizing
// work
type WorkerPool[T any] struct {
	preworkFunc  PrePostWorkFunc
	postworkFunc PrePostWorkFunc
	workFunc     WorkFunc[T]

	poolSize       int
	workChan       chan T
	workChanBuffer int
	wg             *sync.WaitGroup
}

const DefaultPoolSize int = 25

func NewWorkerPool[T any](workFunc WorkFunc[T], options ...PoolOption[T]) (*WorkerPool[T], error) {
	pool := &WorkerPool[T]{
		workFunc:       workFunc,
		poolSize:       DefaultPoolSize,
		workChanBuffer: 1,
		wg:             &sync.WaitGroup{},
	}

	for _, opt := range options {
		if err := opt(pool); err != nil {
			return nil, err
		}
	}

	pool.wg.Add(pool.poolSize)

	return pool, nil
}

// Start will launch all the worker funcs and return two channels. The first channel
// is one that you send work to, and the second one will receive errors from the workers
// if the work function returns an error
func (w *WorkerPool[T]) Start(ctx context.Context) chan T {
	w.workChan = make(chan T, w.workChanBuffer)

	for range w.poolSize {
		go w.worker(ctx)
	}

	return w.workChan
}

// Stop is a convenience method that will close both channels returned by Start and
// wait for all workers to finish.
func (w *WorkerPool[T]) Stop() {
	close(w.workChan)

	w.wg.Wait()
}

func (w *WorkerPool[T]) worker(ctx context.Context) {
	defer w.wg.Done()

	if w.preworkFunc != nil {
		w.preworkFunc(ctx)
	}

	for t := range w.workChan {
		w.workFunc(ctx, t)
	}

	if w.postworkFunc != nil {
		w.postworkFunc(ctx)
	}
}
