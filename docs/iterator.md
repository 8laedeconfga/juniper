# `package iterator`

```
import "github.com/bradenaw/juniper/iterator"
```

# Overview

package iterator allows iterating over sequences of values, for example the contents of a
container.


# Index

<pre><a href="#Collect">func Collect[T any](iter Iterator[T]) []T</a></pre>
<pre><a href="#Equal">func Equal[T comparable](iters ...Iterator[T]) bool</a></pre>
<pre><a href="#Last">func Last[T any](iter Iterator[T], n int) []T</a></pre>
<pre><a href="#Reduce">func Reduce[T any, U any](iter Iterator[T], initial U, f func(U, T) U) U</a></pre>
<pre><a href="#Iterator">type Iterator</a></pre>
<pre>    <a href="#Chan">func Chan[T any](c &lt;-chan T) Iterator[T]</a></pre>
<pre>    <a href="#Chunk">func Chunk[T any](iter Iterator[T], chunkSize int) Iterator[[]T]</a></pre>
<pre>    <a href="#Compact">func Compact[T comparable](iter Iterator[T]) Iterator[T]</a></pre>
<pre>    <a href="#CompactFunc">func CompactFunc[T any](iter Iterator[T], eq func(T, T) bool) Iterator[T]</a></pre>
<pre>    <a href="#Counter">func Counter(n int) Iterator[int]</a></pre>
<pre>    <a href="#Filter">func Filter[T any](iter Iterator[T], keep func(T) bool) Iterator[T]</a></pre>
<pre>    <a href="#First">func First[T any](iter Iterator[T], n int) Iterator[T]</a></pre>
<pre>    <a href="#Flatten">func Flatten[T any](iter Iterator[Iterator[T]]) Iterator[T]</a></pre>
<pre>    <a href="#Join">func Join[T any](iters ...Iterator[T]) Iterator[T]</a></pre>
<pre>    <a href="#Map">func Map[T any, U any](iter Iterator[T], f func(t T) U) Iterator[U]</a></pre>
<pre>    <a href="#Repeat">func Repeat[T any](item T, n int) Iterator[T]</a></pre>
<pre>    <a href="#Runs">func Runs[T any](iter Iterator[T], same func(a, b T) bool) Iterator[Iterator[T]]</a></pre>
<pre>    <a href="#Slice">func Slice[T any](s []T) Iterator[T]</a></pre>
<pre>    <a href="#While">func While[T any](iter Iterator[T], f func(T) bool) Iterator[T]</a></pre>
<pre><a href="#Peekable">type Peekable</a></pre>
<pre>    <a href="#WithPeek">func WithPeek[T any](iter Iterator[T]) Peekable[T]</a></pre>

# Constants

This section is empty.

# Variables

This section is empty.

# Functions

## <a id="Collect"></a><pre>func <a href="#Collect">Collect</a>[T any](iter <a href="#Iterator">Iterator</a>[T]) []T</pre>

Collect advances iter to the end and returns all of the items seen as a slice.


## <a id="Equal"></a><pre>func <a href="#Equal">Equal</a>[T comparable](iters ...) bool</pre>

Equal returns true if the given iterators yield the same items in the same order. Consumes the
iterators.


### Example 
```go
{
	fmt.Println(
		iterator.Equal(
			iterator.Slice([]string{"a", "b", "c"}),
			iterator.Slice([]string{"a", "b", "c"}),
		),
	)

	fmt.Println(
		iterator.Equal(
			iterator.Slice([]string{"a", "b", "c"}),
			iterator.Slice([]string{"a", "b", "c", "d"}),
		),
	)

}
```

Output:
```text
true
false

```

## <a id="Last"></a><pre>func <a href="#Last">Last</a>[T any](iter <a href="#Iterator">Iterator</a>[T], n int) []T</pre>

Last consumes iter and returns the last n items. If iter yields fewer than n items, Last returns
all of them.


### Example 
```go
{
	iter := iterator.Counter(10)

	last3 := iterator.Last(iter, 3)
	fmt.Println(last3)

	iter = iterator.Counter(2)
	last3 = iterator.Last(iter, 3)
	fmt.Println(last3)

}
```

Output:
```text
[7 8 9]
[0 1]

```

## <a id="Reduce"></a><pre>func <a href="#Reduce">Reduce</a>[T any, U any](iter <a href="#Iterator">Iterator</a>[T], initial U, f (U, T) U) U</pre>

Reduce reduces iter to a single value using the reduction function f.


### Example 
```go
{
	x := []int{3, 1, 2}

	iter := iterator.Slice(x)
	sum := iterator.Reduce(iter, 0, func(x, y int) int { return x + y })
	fmt.Println(sum)

	iter = iterator.Slice(x)
	min := iterator.Reduce(iter, math.MaxInt, xmath.Min[int])
	fmt.Println(min)

}
```

Output:
```text
6
1

```

# Types

## <a id="Iterator"></a><pre>type Iterator</pre>
```go
type Iterator[T any] interface {
	// Next advances the iterator and returns the next item. Once the iterator is finished, the
	// first return is meaningless and the second return is false. Note that the final value of the
	// iterator has true in the second return, and it's the following call that returns false in the
	// second return.
	Next() (T, bool)
}
```

Iterator is used to iterate over a sequence of values.

Iterators are lazy, meaning they do no work until a call to Next().

Iterators do not need to be fully consumed, callers may safely abandon an iterator before Next
returns false.


## <a id="Chan"></a><pre>func Chan[T any](c &lt;-chan T) <a href="#Iterator">Iterator</a>[T]</pre>

Chan returns an Iterator that yields the values received on c.


## <a id="Chunk"></a><pre>func Chunk[T any](iter <a href="#Iterator">Iterator</a>[T], chunkSize int) <a href="#Iterator">Iterator</a>[[]T]</pre>

Chunk returns an iterator over non-overlapping chunks of size chunkSize. The last chunk will be
smaller than chunkSize if the iterator does not contain an even multiple.


### Example 
```go
{
	iter := iterator.Slice([]string{"a", "b", "c", "d", "e", "f", "g", "h"})

	chunked := iterator.Chunk(iter, 3)
	item, _ := chunked.Next()
	fmt.Println(item)
	item, _ = chunked.Next()
	fmt.Println(item)
	item, _ = chunked.Next()
	fmt.Println(item)

}
```

Output:
```text
[a b c]
[d e f]
[g h]

```

## <a id="Compact"></a><pre>func Compact[T comparable](iter <a href="#Iterator">Iterator</a>[T]) <a href="#Iterator">Iterator</a>[T]</pre>

Compact elides adjacent duplicates from iter.


### Example 
```go
{
	iter := iterator.Slice([]string{"a", "a", "b", "c", "c", "c", "a"})
	compacted := iterator.Compact(iter)
	fmt.Println(iterator.Collect(compacted))

}
```

Output:
```text
[a b c a]

```

## <a id="CompactFunc"></a><pre>func CompactFunc[T any](iter <a href="#Iterator">Iterator</a>[T], eq (T, T) bool) <a href="#Iterator">Iterator</a>[T]</pre>

CompactFunc elides adjacent duplicates from iter, using eq to determine duplicates.


### Example 
```go
{
	iter := iterator.Slice([]string{
		"bank",
		"beach",
		"ghost",
		"goat",
		"group",
		"yaw",
		"yew",
	})
	compacted := iterator.CompactFunc(iter, func(a, b string) bool {
		return a[0] == b[0]
	})
	fmt.Println(iterator.Collect(compacted))

}
```

Output:
```text
[bank ghost yaw]

```

## <a id="Counter"></a><pre>func Counter(n int) <a href="#Iterator">Iterator</a>[int]</pre>

Counter returns an iterator that counts up from 0, yielding n items.

The following are equivalent:

  for i := 0; i < n; i++ {
    fmt.Println(n)
  }

  iter := iterator.Counter(n)
  for {
    item, ok := iter.Next()
    if !ok {
      break
    }
    fmt.Println(item)
  }


## <a id="Filter"></a><pre>func Filter[T any](iter <a href="#Iterator">Iterator</a>[T], keep (T) bool) <a href="#Iterator">Iterator</a>[T]</pre>

Filter returns an iterator that yields only the items from iter for which keep returns true.


### Example 
```go
{
	iter := iterator.Slice([]int{1, 2, 3, 4, 5, 6})

	evens := iterator.Filter(iter, func(x int) bool { return x%2 == 0 })
	fmt.Println(iterator.Collect(evens))

}
```

Output:
```text
[2 4 6]

```

## <a id="First"></a><pre>func First[T any](iter <a href="#Iterator">Iterator</a>[T], n int) <a href="#Iterator">Iterator</a>[T]</pre>

First returns an iterator that yields the first n items from iter.


### Example 
```go
{
	iter := iterator.Slice([]string{"a", "b", "c", "d", "e"})

	first3 := iterator.First(iter, 3)
	fmt.Println(iterator.Collect(first3))

}
```

Output:
```text
[a b c]

```

## <a id="Flatten"></a><pre>func Flatten[T any](iter <a href="#Iterator">Iterator</a>[<a href="#Iterator">Iterator</a>[T]]) <a href="#Iterator">Iterator</a>[T]</pre>

Flatten returns an iterator that yields all items from all iterators yielded by iter.


### Example 
```go
{
	iter := iterator.Slice([]iterator.Iterator[int]{
		iterator.Slice([]int{0, 1, 2}),
		iterator.Slice([]int{3, 4, 5, 6}),
		iterator.Slice([]int{7}),
	})

	all := iterator.Flatten(iter)

	fmt.Println(iterator.Collect(all))

}
```

Output:
```text
[0 1 2 3 4 5 6 7]

```

## <a id="Join"></a><pre>func Join[T any](iters ...) <a href="#Iterator">Iterator</a>[T]</pre>

Join returns an Iterator that returns all elements of iters[0], then all elements of iters[1],
and so on.


### Example 
```go
{
	iter := iterator.Join(
		iterator.Counter(3),
		iterator.Counter(5),
		iterator.Counter(2),
	)

	fmt.Println(iterator.Collect(iter))

}
```

Output:
```text
[0 1 2 0 1 2 3 4 0 1]

```

## <a id="Map"></a><pre>func Map[T any, U any](iter <a href="#Iterator">Iterator</a>[T], f (t T) U) <a href="#Iterator">Iterator</a>[U]</pre>

Map transforms the results of iter using the conversion f.


## <a id="Repeat"></a><pre>func Repeat[T any](item T, n int) <a href="#Iterator">Iterator</a>[T]</pre>

Repeat returns an iterator that yields item n times.


### Example 
```go
{
	iter := iterator.Repeat("a", 4)
	fmt.Println(iterator.Collect(iter))

}
```

Output:
```text
[a a a a]

```

## <a id="Runs"></a><pre>func Runs[T any](iter <a href="#Iterator">Iterator</a>[T], same (a, b T) bool) <a href="#Iterator">Iterator</a>[<a href="#Iterator">Iterator</a>[T]]</pre>

Runs returns an iterator of iterators. The inner iterators yield contiguous elements from iter
such that same(a, b) returns true for any a and b in the run.

The inner iterator should be drained before calling Next on the outer iterator.

same(a, a) must return true. If same(a, b) and same(b, c) both return true, then same(a, c) must
also.


### Example 
```go
{
	iter := iterator.Slice([]int{2, 4, 0, 7, 1, 3, 9, 2, 8})

	parityRuns := iterator.Runs(iter, func(a, b int) bool {
		return a%2 == b%2
	})
	fmt.Println(iterator.Collect(iterator.Map(parityRuns, iterator.Collect[int])))

}
```

Output:
```text
[[2 4 0] [7 1 3 9] [2 8]]

```

## <a id="Slice"></a><pre>func Slice[T any](s []T) <a href="#Iterator">Iterator</a>[T]</pre>

Slice returns an iterator over the elements of s.


## <a id="While"></a><pre>func While[T any](iter <a href="#Iterator">Iterator</a>[T], f (T) bool) <a href="#Iterator">Iterator</a>[T]</pre>

While returns an iterator that terminates before the first item from iter for which f returns
false.


## <a id="Peekable"></a><pre>type Peekable</pre>
```go
type Peekable[T any] interface {
	Iterator[T]
	// Peek returns the next item of the iterator if there is one without consuming it.
	//
	// If Peek returns a value, the next call to Next will return the same value.
	Peek() (T, bool)
}
```

Peekable allows viewing the next item from an iterator without consuming it.


## <a id="WithPeek"></a><pre>func WithPeek[T any](iter <a href="#Iterator">Iterator</a>[T]) <a href="#Peekable">Peekable</a>[T]</pre>

WithPeek returns iter with a Peek() method attached.

