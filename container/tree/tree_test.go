package tree

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bradenaw/xstd/iterator"
	"github.com/bradenaw/xstd/xsort"
)

func TestBasic(t *testing.T) {
	tree := newTree[int](xsort.OrderedLess[int])

	_, ok := tree.Get(5)
	require.False(t, ok)
	tree.Put(5)
	_, ok = tree.Get(5)
	require.True(t, ok)
}

func FuzzBasic(f *testing.F) {
	const (
		ActPut byte = iota << 6
		ActDelete
		ActContains
		ActCheck
	)

	f.Add([]byte{
		ActContains & 5,
		ActPut & 5,
		ActContains & 5,
		ActDelete & 5,
		ActContains & 5,
	})

	f.Fuzz(func(t *testing.T, b []byte) {
		tree := newTree(xsort.OrderedLess[byte])
		oracle := make(map[byte]struct{})
		for i := range b {
			item := b[i] & 0b00111111
			switch b[i] & 0b11000000 {
			case ActPut:
				tree.Put(item)
				oracle[item] = struct{}{}
			case ActDelete:
				tree.Delete(item)
				delete(oracle, item)
			case ActContains:
				_, treeOk := tree.Get(item)
				_, oracleOk := oracle[item]
				require.Equal(t, treeOk, oracleOk)
				require.Equal(t, tree.Contains(item), oracleOk)
			case ActCheck:
				require.Equal(t, tree.size, len(oracle))
			default:
				panic("no action?")
			}

			var oracleSlice []byte
			for item := range oracle {
				oracleSlice = append(oracleSlice, item)
			}
			xsort.Slice(oracleSlice, xsort.OrderedLess[byte])
			treeSlice := iterator.Collect(tree.Iterate())
			require.Equal(t, treeSlice, oracleSlice)
		}
	})
}