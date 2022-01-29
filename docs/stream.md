# `package stream`

```
import "github.com/bradenaw/juniper/stream"
```

# Overview

package stream allows iterating over sequences of values where iteration may fail.


# Index

<pre><a href="#Collect">func Collect[T any](ctx context.Context, s Stream[T]) ([]T, error)</a></pre>
<pre><a href="#Last">func Last[T any](ctx context.Context, s Stream[T], n int) ([]T, error)</a></pre>
<pre><a href="#Pipe">func Pipe[T any](bufferSize int) (*PipeSender[T], Stream[T])</a></pre>
<pre><a href="#Reduce">func Reduce[T any, U any](
	ctx context.Context,
	s Stream[T],
	initial U,
	f func(U, T) (U, error),
) (U, error)</a></pre>
<pre><a href="#Peekable">type Peekable</a></pre>
<pre>    <a href="#WithPeek">func WithPeek[T any](s Stream[T]) Peekable[T]</a></pre>
<pre><a href="#PipeSender">type PipeSender</a></pre>
<pre>    <a href="#Close">func (s *PipeSender[T]) Close(err error)</a></pre>
<pre>    <a href="#Send">func (s *PipeSender[T]) Send(ctx context.Context, x T) error</a></pre>
<pre><a href="#Stream">type Stream</a></pre>
<pre>    <a href="#Batch">func Batch[T any](s Stream[T], maxWait time.Duration, batchSize int) Stream[[]T]</a></pre>
<pre>    <a href="#BatchFunc">func BatchFunc[T any](
    	s Stream[T],
    	maxWait time.Duration,
    	full func(batch []T) bool,
    ) Stream[[]T]</a></pre>
<pre>    <a href="#Chan">func Chan[T any](c &lt;-chan T) Stream[T]</a></pre>
<pre>    <a href="#Chunk">func Chunk[T any](s Stream[T], chunkSize int) Stream[[]T]</a></pre>
<pre>    <a href="#Compact">func Compact[T comparable](s Stream[T]) Stream[T]</a></pre>
<pre>    <a href="#CompactFunc">func CompactFunc[T comparable](s Stream[T], eq func(T, T) bool) Stream[T]</a></pre>
<pre>    <a href="#Filter">func Filter[T any](s Stream[T], keep func(T) (bool, error)) Stream[T]</a></pre>
<pre>    <a href="#First">func First[T any](s Stream[T], n int) Stream[T]</a></pre>
<pre>    <a href="#Flatten">func Flatten[T any](s Stream[Stream[T]]) Stream[T]</a></pre>
<pre>    <a href="#FromIterator">func FromIterator[T any](iter iterator.Iterator[T]) Stream[T]</a></pre>
<pre>    <a href="#Join">func Join[T any](streams ...Stream[T]) Stream[T]</a></pre>
<pre>    <a href="#Map">func Map[T any, U any](s Stream[T], f func(t T) (U, error)) Stream[U]</a></pre>
<pre>    <a href="#Runs">func Runs[T any](s Stream[T], same func(a, b T) bool) Stream[Stream[T]]</a></pre>
<pre>    <a href="#While">func While[T any](s Stream[T], f func(T) (bool, error)) Stream[T]</a></pre>

# Constants

This section is empty.

# Variables

<pre>
<a id="ErrClosedPipe"></a><a id="End"></a>var (
    ErrClosedPipe = errors.New("closed pipe")
    End = errors.New("end of stream")
)
</pre>

# Functions

## <a id="Collect"></a><pre>func <a href="#Collect">Collect</a>[T any](ctx <a href="https://pkg.go.dev/context#Context">context.Context</a>, s <a href="#Stream">Stream</a>[T]) ([]T, error)</pre>

Collect advances s to the end and returns all of the items seen as a slice.


### Example 
```go
{
	ctx := context.Background()
	s := stream.FromIterator(iterator.Slice([]string{"a", "b", "c"}))

	x, err := stream.Collect(ctx, s)
	fmt.Println(err)
	fmt.Println(x)

}
```

Output:
```text
<nil>
[a b c]

```

## <a id="Last"></a><pre>func <a href="#Last">Last</a>[T any](ctx <a href="https://pkg.go.dev/context#Context">context.Context</a>, s <a href="#Stream">Stream</a>[T], n int) ([]T, error)</pre>

Last consumes s and returns the last n items. If s yields fewer than n items, Last returns
all of them.


## <a id="Pipe"></a><pre>func <a href="#Pipe">Pipe</a>[T any](bufferSize int) (*<a href="#PipeSender">PipeSender</a>[T], <a href="#Stream">Stream</a>[T])</pre>

Pipe returns a linked sender and receiver pair. Values sent using sender.Send will be delivered
to the given Stream. The Stream will terminate when the sender is closed.

bufferSize is the number of elements in the buffer between the sender and the receiver. 0 has the
same meaning as for the built-in make(chan).


### Example 
```go
{
	ctx := context.Background()
	sender, receiver := stream.Pipe[int](0)

	go func() {
		sender.Send(ctx, 1)
		sender.Send(ctx, 2)
		sender.Send(ctx, 3)
		sender.Close(nil)
	}()

	defer receiver.Close()
	for {
		item, err := receiver.Next(ctx)
		if err == stream.End {
			break
		} else if err != nil {
			fmt.Printf("stream ended with error: %s\n", err)
			return
		}
		fmt.Println(item)
	}

}
```

Output:
```text
1
2
3

```

### Example error
```go
{
	ctx := context.Background()
	sender, receiver := stream.Pipe[int](0)

	oopsError := errors.New("oops")

	go func() {
		sender.Send(ctx, 1)
		sender.Close(oopsError)
	}()

	defer receiver.Close()
	for {
		item, err := receiver.Next(ctx)
		if err == stream.End {
			fmt.Println("stream ended normally")
			break
		} else if err != nil {
			fmt.Printf("stream ended with error: %s\n", err)
			return
		}
		fmt.Println(item)
	}

}
```

Output:
```text
1
stream ended with error: oops

```

## <a id="Reduce"></a><pre>func <a href="#Reduce">Reduce</a>[T any, U any](ctx <a href="https://pkg.go.dev/context#Context">context.Context</a>, s <a href="#Stream">Stream</a>[T], initial U, f (U, T) (U, error)) (U, error)</pre>

Reduce reduces s to a single value using the reduction function f.


# Types

## <a id="Peekable"></a><pre>type Peekable</pre>
```go
type Peekable[T any] interface {
	Stream[T]
	// Peek returns the next item of the stream if there is one without consuming it.
	//
	// If Peek returns a value, the next call to Next will return the same value.
	Peek(ctx context.Context) (T, error)
}
```

Peekable allows viewing the next item from a stream without consuming it.


## <a id="WithPeek"></a><pre>func WithPeek[T any](s <a href="#Stream">Stream</a>[T]) <a href="#Peekable">Peekable</a>[T]</pre>

WithPeek returns iter with a Peek() method attached.


## <a id="PipeSender"></a><pre>type PipeSender</pre>
```go
type PipeSender[T any] struct {
	// contains filtered or unexported fields
}
```

PipeSender is the send half of a pipe returned by Pipe.


## <a id="Close"></a><pre>func (s *<a href="#PipeSender">PipeSender</a>[T]) Close(err error)</pre>

Close closes the PipeSender, signalling to the receiver that no more values will be sent. If an
error is provided, it will surface when closing the receiver.


## <a id="Send"></a><pre>func (s *<a href="#PipeSender">PipeSender</a>[T]) Send(ctx <a href="https://pkg.go.dev/context#Context">context.Context</a>, x T) error</pre>

Send attempts to send x to the receiver. If the receiver closes before x can be sent, returns
ErrClosedPipe immediately. If ctx expires before x can be sent, returns ctx.Err().

A nil return does not necessarily mean that the receiver will see x, since the receiver may close
early.


## <a id="Stream"></a><pre>type Stream</pre>
```go
type Stream[T any] interface {
	// Next advances the stream and returns the next item. If the stream is already over, Next
	// returns stream.End in the second return. Note that the final item of the stream has nil in
	// the second return, and it's the following call that returns stream.End.
	Next(ctx context.Context) (T, error)
	// Close ends receiving from the stream. It is invalid to call Next after calling Close.
	Close()
}
```

Stream is used to iterate over a sequence of values. It is similar to Iterator, except intended
for use when iteration may fail for some reason, usually because the sequence requires I/O to
produce.

Streams and the combinator functions are lazy, meaning they do no work until a call to Next().

Streams do not need to be fully consumed, but streams must be closed. Functions in this package
that are passed streams expect to be the sole user of that stream going forward, and so will
handle closing on your behalf so long as all streams they return are closed appropriately.


## <a id="Batch"></a><pre>func Batch[T any](s <a href="#Stream">Stream</a>[T], maxWait <a href="https://pkg.go.dev/time#Duration">time.Duration</a>, batchSize int) <a href="#Stream">Stream</a>[[]T]</pre>

Batch returns a stream of non-overlapping batches from s of size batchSize. Batch is similar to
Chunk with the added feature that an underfilled batch will be delivered to the output stream if
any item has been in the batch for more than maxWait.


### Example 
```go
{
	ctx := context.Background()

	sender, receiver := stream.Pipe[string](0)
	batchStream := stream.Batch(receiver, 50*time.Millisecond, 3)

	wait := make(chan struct{}, 3)
	go func() {
		_ = sender.Send(ctx, "a")
		_ = sender.Send(ctx, "b")

		<-wait
		_ = sender.Send(ctx, "c")
		_ = sender.Send(ctx, "d")
		_ = sender.Send(ctx, "e")
		_ = sender.Send(ctx, "f")
		sender.Close(nil)
	}()

	defer batchStream.Close()
	var batches [][]string
	for {
		batch, err := batchStream.Next(ctx)
		if err == stream.End {
			break
		} else if err != nil {
			fmt.Printf("stream ended with error: %s\n", err)
			return
		}
		batches = append(batches, batch)
		wait <- struct{}{}
	}
	fmt.Println(batches)

}
```

Output:
```text
[[a b] [c d e] [f]]

```

## <a id="BatchFunc"></a><pre>func BatchFunc[T any](s <a href="#Stream">Stream</a>[T], maxWait <a href="https://pkg.go.dev/time#Duration">time.Duration</a>, full (batch []T) bool) <a href="#Stream">Stream</a>[[]T]</pre>

BatchFunc returns a stream of non-overlapping batches from s, using full to determine when a
batch is full. BatchFunc is similar to Chunk with the added feature that an underfilled batch
will be delivered to the output stream if any item has been in the batch for more than maxWait.


## <a id="Chan"></a><pre>func Chan[T any](c &lt;-chan T) <a href="#Stream">Stream</a>[T]</pre>

Chan returns a Stream that receives values from c.


## <a id="Chunk"></a><pre>func Chunk[T any](s <a href="#Stream">Stream</a>[T], chunkSize int) <a href="#Stream">Stream</a>[[]T]</pre>

Chunk returns a stream of non-overlapping chunks from s of size chunkSize. The last chunk will be
smaller than chunkSize if the stream does not contain an even multiple.


## <a id="Compact"></a><pre>func Compact[T comparable](s <a href="#Stream">Stream</a>[T]) <a href="#Stream">Stream</a>[T]</pre>

Compact elides adjacent duplicates from s.


## <a id="CompactFunc"></a><pre>func CompactFunc[T comparable](s <a href="#Stream">Stream</a>[T], eq (T, T) bool) <a href="#Stream">Stream</a>[T]</pre>

CompactFunc elides adjacent duplicates from s, using eq to determine duplicates.


## <a id="Filter"></a><pre>func Filter[T any](s <a href="#Stream">Stream</a>[T], keep (T) (bool, error)) <a href="#Stream">Stream</a>[T]</pre>

Filter returns a Stream that yields only the items from s for which keep returns true. If keep
returns an error, terminates the stream early.


## <a id="First"></a><pre>func First[T any](s <a href="#Stream">Stream</a>[T], n int) <a href="#Stream">Stream</a>[T]</pre>

First returns a Stream that yields the first n items from s.


## <a id="Flatten"></a><pre>func Flatten[T any](s <a href="#Stream">Stream</a>[<a href="#Stream">Stream</a>[T]]) <a href="#Stream">Stream</a>[T]</pre>

Flatten returns a stream that yields all items from all streams yielded by s.


## <a id="FromIterator"></a><pre>func FromIterator[T any](iter <a href="./iterator.md#Iterator">iterator.Iterator</a>[T]) <a href="#Stream">Stream</a>[T]</pre>

FromIterator returns a Stream that yields the values from iter. This stream ignores the context
passed to Next during the call to iter.Next.


## <a id="Join"></a><pre>func Join[T any](streams ...) <a href="#Stream">Stream</a>[T]</pre>

Join returns a Stream that yields all elements from streams[0], then all elements from
streams[1], and so on.


## <a id="Map"></a><pre>func Map[T any, U any](s <a href="#Stream">Stream</a>[T], f (t T) (U, error)) <a href="#Stream">Stream</a>[U]</pre>

Map transforms the values of s using the conversion f. If f returns an error, terminates the
stream early.


## <a id="Runs"></a><pre>func Runs[T any](s <a href="#Stream">Stream</a>[T], same (a, b T) bool) <a href="#Stream">Stream</a>[<a href="#Stream">Stream</a>[T]]</pre>

Runs returns a stream of streams. The inner streams yield contiguous elements from s such that
same(a, b) returns true for any a and b in the run.

The inner stream should be drained before calling Next on the outer stream.

same(a, a) must return true. If same(a, b) and same(b, c) both return true, then same(a, c) must
also.


## <a id="While"></a><pre>func While[T any](s <a href="#Stream">Stream</a>[T], f (T) (bool, error)) <a href="#Stream">Stream</a>[T]</pre>

While returns a Stream that terminates before the first item from s for which f returns false.
If f returns an error, terminates the stream early.

