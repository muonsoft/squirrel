package squirrel

import (
	"strconv"
	"testing"
)

var benchmarkSlice []int
var benchmarkList immutableList[int]

// BenchmarkCollections isolates the representation choice from SQL rendering.
func BenchmarkCollections(b *testing.B) {
	for _, length := range []int{4, 32, 256} {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			b.Run("CopyOnWrite", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					var values []int
					for i := range length {
						values = append(values[:len(values):len(values)], i)
					}
					benchmarkSlice = values
				}
			})
			b.Run("Persistent", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					var values immutableList[int]
					for i := range length {
						values = appendPersistent(values, i)
					}
					benchmarkList = values
				}
			})
		})
	}
}
