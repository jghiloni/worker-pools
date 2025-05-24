package pool

import (
	"fmt"
	"math"
)

// PoolOptions are passed to NewWorkerPool to modify the pool as it's being built
type PoolOption[T any] func(wp *WorkerPool[T]) error

// WithPreworkFunc will set the func that will be called when a pool worker starts.
// If it is nil, it will not be called.
func WithPreworkFunc[T any](f PrePostWorkFunc) PoolOption[T] {
	return func(wp *WorkerPool[T]) error {
		wp.preworkFunc = f
		return nil
	}
}

// WithPreworkFunc will set the func that will be called when a pool worker ends --
// that is, when the work channel is closed. If it is nil, it will not be called.
func WithPostWorkFunc[T any](f PrePostWorkFunc) PoolOption[T] {
	return func(wp *WorkerPool[T]) error {
		wp.postworkFunc = f
		return nil
	}
}

// WithPoolSize will set the number of workers simultaneously working in the pool.
// Obviously, your system resources will dictate how big your pool can be and your
// use case will dictate how big it SHOULD be. By default, 25 workers will be created.
// If a number < 1 is passed, an error will be returned
func WithPoolSize[T any](size int) PoolOption[T] {
	return func(wp *WorkerPool[T]) error {
		if size < 1 {
			return fmt.Errorf("pool size %d cannot be less than 1", size)
		}

		wp.poolSize = size
		return nil
	}
}

// WithWorkChannelBuffer will create a work channel with a buffer of n items
// or unbuffered, if n == 0. By default, it is set to 1 item. If n > MaxInt,
// an error is returned
func WithWorkChannelBuffer[T any](n uint64) PoolOption[T] {
	return func(wp *WorkerPool[T]) error {
		if n > math.MaxInt {
			return fmt.Errorf("buffer cannot be larger than %d", math.MaxInt)
		}

		wp.workChanBuffer = int(n)
		return nil
	}
}
