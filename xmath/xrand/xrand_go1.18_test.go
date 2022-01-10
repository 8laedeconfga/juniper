//go:build go1.18

package xrand

import (
	"context"
	"encoding/binary"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/stream"
)

func FuzzSampleInner(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte, k int) {
		if k <= 0 {
			return
		}
		randIntn := func(n int) int {
			if len(b) < 4 {
				return 0
			}
			x := binary.BigEndian.Uint32(b[:4])
			b = b[4:]
			return int(x) % n
		}
		randFloat64 := func() float64 {
			if len(b) < 8 {
				return 0
			}
			x := binary.BigEndian.Uint64(b[:8])
			b = b[8:]
			out := float64(x) / math.MaxUint64
			if out == 1 {
				out = math.Nextafter(out, 0)
			}
			require.GreaterOrEqual(t, out, float64(0))
			require.Less(t, out, float64(1))
			t.Logf("%f", out)
			return out
		}

		t.Logf("k %d", k)

		sampler := sampleInner(randFloat64, randIntn, k)
		prev := 0
		for i := 0; i < 100; i++ {
			next, replace := sampler()
			t.Logf("%d: next %d replace %d", i, next, replace)
			if next == math.MaxInt {
				break
			}
			if i < k {
				require.Equal(t, next, i)
				require.Equal(t, replace, i)
			} else {
				require.Greater(t, next, prev)
				require.GreaterOrEqual(t, replace, 0)
				require.Less(t, replace, k)
			}

			prev = next
		}
	})
}

func stddev(a []int) float64 {
	m := mean(a)
	sumSquaredDeviation := float64(0)
	for i := range a {
		deviation := m - float64(a[i])
		sumSquaredDeviation += (deviation * deviation)
	}
	return math.Sqrt(sumSquaredDeviation / float64(len(a)))
}

func mean(a []int) float64 {
	sum := 0
	for i := range a {
		sum += a[i]
	}
	return float64(sum) / float64(len(a))
}

// f must return the same as Sample(r, 20, 5).
func testSample(t *testing.T, f func(r *rand.Rand) []int) {
	r := rand.New(rand.NewSource(0))

	counts := make([]int, 20)

	for i := 0; i < 10000; i++ {
		sample := f(r)
		for _, item := range sample {
			counts[item]++
		}
	}
	m := mean(counts)

	t.Logf("counts        %#v", counts)
	t.Logf("stddev        %#v", stddev(counts))
	t.Logf("stddev / mean %#v", stddev(counts)/m)

	// There's certainly a better statistical test than this, but I haven't bothered to break out
	// the stats book yet.
	require.InDelta(t, 0.02, stddev(counts)/m, 0.01)

}
func TestSample(t *testing.T) {
	testSample(t, func(r *rand.Rand) []int {
		return Sample(r, 20, 5)
	})
}

func TestSampleSlice(t *testing.T) {
	a := iterator.Collect(iterator.Counter(20))
	testSample(t, func(r *rand.Rand) []int {
		return SampleSlice(r, a, 5)
	})
}

func TestSampleIterator(t *testing.T) {
	testSample(t, func(r *rand.Rand) []int {
		return SampleIterator(r, iterator.Counter(20), 5)
	})
}

func TestSampleStream(t *testing.T) {
	testSample(t, func(r *rand.Rand) []int {
		out, err := SampleStream(
			context.Background(),
			r,
			stream.FromIterator(iterator.Counter(20)),
			5,
		)
		require.NoError(t, err)
		return out
	})
}