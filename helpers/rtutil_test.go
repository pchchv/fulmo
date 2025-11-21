package helpers

import (
	crand "crypto/rand"
	"testing"
)

func BenchmarkMemHash(b *testing.B) {
	buf := make([]byte, 64)
	crand.Read(buf)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MemHash(buf)
	}

	b.SetBytes(int64(len(buf)))
}

func benchmarkRand(b *testing.B, fab func() func() uint32) {
	b.RunParallel(func(pb *testing.PB) {
		gen := fab()
		for pb.Next() {
			gen()
		}
	})
}
