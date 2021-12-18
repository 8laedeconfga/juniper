package parallel

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

// Do calls f from parallelism goroutines n times, providing each invocation a unique i in [0, n).
//
// If parallelism <= 0, uses GOMAXPROCS instead.
func Do(
	parallelism int,
	n int,
	f func(i int),
) {
	if parallelism <= 0 {
		parallelism = runtime.GOMAXPROCS(-1)
	}

	if parallelism == 1 {
		for i := 0; i < n; i++ {
			f(i)
		}
		return
	}

	x := int32(-1)
	var wg sync.WaitGroup
	wg.Add(parallelism)
	for j := 0; j < parallelism; j++ {
		go func() {
			defer wg.Done()
			for {
				i := int(atomic.AddInt32(&x, 1))
				if i >= n {
					return
				}
				f(i)
			}
		}()
	}
	wg.Wait()
	return
}

// DoContext calls f from parallelism goroutines n times, providing each invocation a unique i in
// [0, n).
//
// If any call to f returns an error the context passed to invocations of f is cancelled, no further
// calls to f are made, and Do returns the first error encountered.
//
// If parallelism <= 0, uses GOMAXPROCS instead.
func DoContext(
	ctx context.Context,
	parallelism int,
	n int,
	f func(ctx context.Context, i int) error,
) error {
	if parallelism <= 0 {
		parallelism = runtime.GOMAXPROCS(-1)
	}

	if parallelism == 1 {
		for i := 0; i < n; i++ {
			err := f(ctx, i)
			if err != nil {
				return err
			}
		}
		return nil
	}

	x := int32(-1)
	eg, ctx := errgroup.WithContext(ctx)
	for j := 0; j < parallelism; j++ {
		eg.Go(func() error {
			for {
				i := int(atomic.AddInt32(&x, 1))
				if i >= n {
					return nil
				}

				if ctx.Err() != nil {
					return ctx.Err()
				}

				err := f(ctx, i)
				if err != nil {
					return err
				}
			}
		})
	}
	return eg.Wait()
}
