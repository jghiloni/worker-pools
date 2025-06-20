package pool_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jghiloni/worker-pools/pool"
)

func prework(counter *int32) pool.PrePostWorkFunc {
	return func(ctx context.Context) {
		atomic.AddInt32(counter, 1)
	}
}

func work(m *sync.Mutex, w io.Writer) pool.WorkFunc[int] {
	return func(ctx context.Context, i int) {
		m.Lock()
		fmt.Fprintln(w, "item", i)
		m.Unlock()

	}
}

func TestPoolWithDifferentBuffers(t *testing.T) {
	expectedPoolSize := 50

	helper := func(bufferSize int) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			actualPoolSize := new(int32)
			buffer := &strings.Builder{}
			m := &sync.Mutex{}

			workPool, err := pool.NewWorkerPool(work(m, buffer),
				pool.WithPoolSize[int](expectedPoolSize),
				pool.WithPreworkFunc[int](prework(actualPoolSize)),
				pool.WithWorkChannelBuffer[int](uint64(bufferSize)),
			)

			if err != nil {
				t.Fatal(err)
			}

			wc := workPool.Start(context.Background())

			for i := range 1000 {
				wc <- i
			}

			for len(wc) > 0 {
				time.Sleep(10 * time.Millisecond)
			}

			workPool.Stop()

			if *actualPoolSize != int32(expectedPoolSize) {
				t.Fatalf("expected pool size %d, but %d were started", expectedPoolSize, *actualPoolSize)
			}

			lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
			if len(lines) != 1000 {
				t.Fatalf("Expected 1000 lines to be printed, but only %d were", len(lines))
			}
		}
	}

	for i := 0; i <= 1000; i += 200 {
		t.Run(fmt.Sprintf("Buffer %d", i), helper(i))
	}
}
