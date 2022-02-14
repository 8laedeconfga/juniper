//go:build go1.18
// +build go1.18

package parallel

import (
	"context"
	"runtime"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/bradenaw/juniper/container/xheap"
	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/stream"
)

// Map uses parallelism goroutines to call f once for each element of in. out[i] is the
// result of f for in[i].
//
// If parallelism <= 0, uses GOMAXPROCS instead.
func Map[T any, U any](
	parallelism int,
	in []T,
	f func(in T) U,
) []U {
	out := make([]U, len(in))
	Do(parallelism, len(in), func(i int) {
		out[i] = f(in[i])
	})
	return out
}

// MapContext uses parallelism goroutines to call f once for each element of in. out[i] is the
// result of f for in[i].
//
// If any call to f returns an error the context passed to invocations of f is cancelled, no further
// calls to f are made, and Map returns the first error encountered.
//
// If parallelism <= 0, uses GOMAXPROCS instead.
func MapContext[T any, U any](
	ctx context.Context,
	parallelism int,
	in []T,
	f func(ctx context.Context, in T) (U, error),
) ([]U, error) {
	out := make([]U, len(in))
	err := DoContext(ctx, parallelism, len(in), func(ctx context.Context, i int) error {
		var err error
		out[i], err = f(ctx, in[i])
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MapIterator uses parallelism goroutines to call f once for each element yielded by iter. The
// returned iterator returns these results in the same order that iter yielded them in.
//
// This iterator, in contrast with most, must be consumed completely or it will leak the goroutines.
//
// If parallelism <= 0, uses GOMAXPROCS instead.
//
// bufferSize is the size of the work buffer. A larger buffer uses more memory but gives better
// throughput in the face of larger variance in the processing time for f.
func MapIterator[T any, U any](
	iter iterator.Iterator[T],
	parallelism int,
	bufferSize int,
	f func(T) U,
) iterator.Iterator[U] {
	if parallelism <= 0 {
		parallelism = runtime.GOMAXPROCS(-1)
	}

	in := make(chan valueAndIndex[T])

	go func() {
		i := 0
		for {
			item, ok := iter.Next()
			if !ok {
				break
			}

			in <- valueAndIndex[T]{
				value: item,
				idx:   i,
			}
			i++
		}
		close(in)
	}()

	c := make(chan valueAndIndex[U], bufferSize)
	nDone := uint32(0)
	for i := 0; i < parallelism; i++ {
		go func() {
			for item := range in {
				u := f(item.value)
				c <- valueAndIndex[U]{value: u, idx: item.idx}
			}
			if atomic.AddUint32(&nDone, 1) == uint32(parallelism) {
				close(c)
			}
		}()
	}

	return &mapIterator[U]{
		c: c,
		h: xheap.New(func(a, b valueAndIndex[U]) bool {
			return a.idx < b.idx
		}, nil),
		i: 0,
	}
}

type mapIterator[U any] struct {
	c <-chan valueAndIndex[U]
	h xheap.Heap[valueAndIndex[U]]
	i int
}

func (iter *mapIterator[U]) Next() (U, bool) {
	for {
		if iter.h.Len() > 0 && iter.h.Peek().idx == iter.i {
			item := iter.h.Pop()
			iter.i++
			return item.value, true
		}
		item, ok := <-iter.c
		if !ok {
			var zero U
			return zero, false
		}
		iter.h.Push(item)
	}
}

type valueAndIndex[T any] struct {
	value T
	idx   int
}

// MapStream uses parallelism goroutines to call f once for each element yielded by s. The returned
// stream returns these results in the same order that s yielded them in.
//
// If any call to f returns an error the context passed to invocations of f is cancelled, no further
// calls to f are made, and the returned stream's Next returns the first error encountered.
//
// If parallelism <= 0, uses GOMAXPROCS instead.
//
// bufferSize is the size of the work buffer. A larger buffer uses more memory but gives better
// throughput in the face of larger variance in the processing time for f.
func MapStream[T any, U any](
	ctx context.Context,
	s stream.Stream[T],
	parallelism int,
	bufferSize int,
	f func(context.Context, T) (U, error),
) stream.Stream[U] {
	if parallelism <= 0 {
		parallelism = runtime.GOMAXPROCS(-1)
	}

	in := make(chan valueAndIndex[T])

	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer s.Close()
		defer close(in)
		i := 0
		for {
			item, err := s.Next(ctx)
			if err == stream.End {
				break
			} else if err != nil {
				return err
			}

			select {
			case in <- valueAndIndex[T]{
				value: item,
				idx:   i,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
			i++
		}
		return nil
	})

	c := make(chan valueAndIndex[U], bufferSize)
	nDone := uint32(0)
	for i := 0; i < parallelism; i++ {
		eg.Go(func() error {
			defer func() {
				if atomic.AddUint32(&nDone, 1) == uint32(parallelism) {
					close(c)
				}
			}()
			for item := range in {
				u, err := f(ctx, item.value)
				if err != nil {
					return err
				}
				select {
				case c <- valueAndIndex[U]{value: u, idx: item.idx}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	return &mapStream[U]{
		cancel: cancel,
		eg:     eg,
		c:      c,
		h: xheap.New(func(a, b valueAndIndex[U]) bool {
			return a.idx < b.idx
		}, nil),
		i: 0,
	}
}

type mapStream[U any] struct {
	cancel context.CancelFunc
	eg     *errgroup.Group
	c      <-chan valueAndIndex[U]
	h      xheap.Heap[valueAndIndex[U]]
	i      int
}

func (s *mapStream[U]) Next(ctx context.Context) (U, error) {
	var zero U
	for {
		if s.h.Len() > 0 && s.h.Peek().idx == s.i {
			item := s.h.Pop()
			s.i++
			return item.value, nil
		}
		select {
		case item, ok := <-s.c:
			if !ok {
				err := s.eg.Wait()
				if err != nil {
					return zero, err
				}
				return zero, stream.End
			}
			s.h.Push(item)
		case <-ctx.Done():
			return zero, ctx.Err()
		}
	}
}

func (s *mapStream[U]) Close() {
	s.cancel()
	_ = s.eg.Wait()
}
