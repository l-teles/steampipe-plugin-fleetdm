package fleetdm

import (
	"context"
	"errors"
	"runtime"
	"sort"
	"sync/atomic"
	"testing"
	"time"
)

func TestFanOut(t *testing.T) {
	t.Parallel()

	inputs := make([]int, 50)
	for i := range inputs {
		inputs[i] = i
	}

	t.Run("all rows, bounded concurrency", func(t *testing.T) {
		t.Parallel()
		var inFlight, peak atomic.Int32
		var got []int
		err := fanOut(context.Background(), inputs, 4, func(_ context.Context, i int) ([]int, error) {
			n := inFlight.Add(1)
			for {
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			inFlight.Add(-1)
			return []int{i, i + 1000}, nil
		}, func(v int) bool {
			got = append(got, v)
			return true
		})
		if err != nil {
			t.Fatalf("fanOut error: %v", err)
		}
		if len(got) != 100 {
			t.Fatalf("got %d rows, want 100", len(got))
		}
		sort.Ints(got)
		if got[0] != 0 || got[49] != 49 || got[50] != 1000 {
			t.Errorf("unexpected rows: %v", got[:3])
		}
		if peak.Load() > 4 {
			t.Errorf("peak concurrency = %d, want <= 4", peak.Load())
		}
	})

	t.Run("error stops and is returned", func(t *testing.T) {
		t.Parallel()
		boom := errors.New("boom")
		err := fanOut(context.Background(), inputs, 4, func(_ context.Context, i int) ([]int, error) {
			if i == 7 {
				return nil, boom
			}
			return []int{i}, nil
		}, func(int) bool { return true })
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v, want boom", err)
		}
	})

	t.Run("emit false stops early without leaking workers", func(t *testing.T) {
		before := runtime.NumGoroutine()
		emitted := 0
		err := fanOut(context.Background(), inputs, 4, func(ctx context.Context, i int) ([]int, error) {
			return []int{i}, ctx.Err()
		}, func(int) bool {
			emitted++
			return emitted < 3
		})
		if err != nil {
			t.Fatalf("fanOut error: %v", err)
		}
		if emitted != 3 {
			t.Errorf("emitted %d rows, want 3", emitted)
		}
		deadline := time.Now().Add(2 * time.Second)
		for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}
		if n := runtime.NumGoroutine(); n > before {
			t.Errorf("goroutines = %d after early stop, want <= %d", n, before)
		}
	})

	t.Run("no inputs", func(t *testing.T) {
		t.Parallel()
		err := fanOut(context.Background(), []int{}, 4, func(context.Context, int) ([]int, error) {
			t.Fatal("fetch called with no inputs")
			return nil, nil
		}, func(int) bool { return true })
		if err != nil {
			t.Fatalf("fanOut error: %v", err)
		}
	})
}
