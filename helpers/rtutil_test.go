package helpers

import (
	crand "crypto/rand"
	"hash/fnv"
	mrand "math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dgryski/go-farm"
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

func BenchmarkMemHashString(b *testing.B) {
	s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MemHashString(s)
	}

	b.SetBytes(int64(len(s)))
}

func BenchmarkSip(b *testing.B) {
	buf := make([]byte, 64)
	crand.Read(buf)
	for i := 0; i < b.N; i++ {
		SipHash(buf)
	}
}

func BenchmarkFnv(b *testing.B) {
	buf := make([]byte, 64)
	crand.Read(buf)
	f := fnv.New64a()
	for i := 0; i < b.N; i++ {
		f.Write(buf)
		f.Sum64()
		f.Reset()
	}
}

func BenchmarkFarm(b *testing.B) {
	buf := make([]byte, 64)
	crand.Read(buf)
	for i := 0; i < b.N; i++ {
		farm.Fingerprint64(buf)
	}
}

func BenchmarkFastRand(b *testing.B) {
	benchmarkRand(b, func() func() uint32 {
		return FastRand
	})
}

func BenchmarkRandSource(b *testing.B) {
	benchmarkRand(b, func() func() uint32 {
		s := mrand.New(mrand.NewSource(time.Now().Unix()))
		return func() uint32 { return s.Uint32() }
	})
}

func BenchmarkRandGlobal(b *testing.B) {
	benchmarkRand(b, func() func() uint32 {
		return func() uint32 { return mrand.Uint32() }
	})
}

func BenchmarkRandAtomic(b *testing.B) {
	var x uint32
	benchmarkRand(b, func() func() uint32 {
		return func() uint32 { return atomic.AddUint32(&x, 1) }
	})
}

func benchmarkRand(b *testing.B, fab func() func() uint32) {
	b.RunParallel(func(pb *testing.PB) {
		gen := fab()
		for pb.Next() {
			gen()
		}
	})
}

func SipHash(p []byte) (l, h uint64) {
	// initialization
	v0 := uint64(8317987320269560794) // k0 ^ 0x736f6d6570736575
	v1 := uint64(7237128889637516672) // k1 ^ 0x646f72616e646f6d
	v2 := uint64(7816392314733513934) // k0 ^ 0x6c7967656e657261
	v3 := uint64(8387220255325274014) // k1 ^ 0x7465646279746573
	t := uint64(len(p)) << 56

	// compression
	for len(p) >= 8 {
		m := uint64(p[0]) | uint64(p[1])<<8 | uint64(p[2])<<16 | uint64(p[3])<<24 |
			uint64(p[4])<<32 | uint64(p[5])<<40 | uint64(p[6])<<48 | uint64(p[7])<<56

		v3 ^= m

		// round 1
		v0 += v1
		v1 = v1<<13 | v1>>51
		v1 ^= v0
		v0 = v0<<32 | v0>>32

		v2 += v3
		v3 = v3<<16 | v3>>48
		v3 ^= v2

		v0 += v3
		v3 = v3<<21 | v3>>43
		v3 ^= v0

		v2 += v1
		v1 = v1<<17 | v1>>47
		v1 ^= v2
		v2 = v2<<32 | v2>>32

		// round 2
		v0 += v1
		v1 = v1<<13 | v1>>51
		v1 ^= v0
		v0 = v0<<32 | v0>>32

		v2 += v3
		v3 = v3<<16 | v3>>48
		v3 ^= v2

		v0 += v3
		v3 = v3<<21 | v3>>43
		v3 ^= v0

		v2 += v1
		v1 = v1<<17 | v1>>47
		v1 ^= v2
		v2 = v2<<32 | v2>>32

		v0 ^= m
		p = p[8:]
	}

	// compress last block
	switch len(p) {
	case 7:
		t |= uint64(p[6]) << 48
		fallthrough
	case 6:
		t |= uint64(p[5]) << 40
		fallthrough
	case 5:
		t |= uint64(p[4]) << 32
		fallthrough
	case 4:
		t |= uint64(p[3]) << 24
		fallthrough
	case 3:
		t |= uint64(p[2]) << 16
		fallthrough
	case 2:
		t |= uint64(p[1]) << 8
		fallthrough
	case 1:
		t |= uint64(p[0])
	}

	v3 ^= t

	// round 1
	v0 += v1
	v1 = v1<<13 | v1>>51
	v1 ^= v0
	v0 = v0<<32 | v0>>32

	v2 += v3
	v3 = v3<<16 | v3>>48
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>43
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>47
	v1 ^= v2
	v2 = v2<<32 | v2>>32

	// round 2
	v0 += v1
	v1 = v1<<13 | v1>>51
	v1 ^= v0
	v0 = v0<<32 | v0>>32

	v2 += v3
	v3 = v3<<16 | v3>>48
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>43
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>47
	v1 ^= v2
	v2 = v2<<32 | v2>>32

	v0 ^= t

	// finalization
	v2 ^= 0xff

	// round 1
	v0 += v1
	v1 = v1<<13 | v1>>51
	v1 ^= v0
	v0 = v0<<32 | v0>>32

	v2 += v3
	v3 = v3<<16 | v3>>48
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>43
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>47
	v1 ^= v2
	v2 = v2<<32 | v2>>32

	// round 2
	v0 += v1
	v1 = v1<<13 | v1>>51
	v1 ^= v0
	v0 = v0<<32 | v0>>32

	v2 += v3
	v3 = v3<<16 | v3>>48
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>43
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>47
	v1 ^= v2
	v2 = v2<<32 | v2>>32

	// round 3
	v0 += v1
	v1 = v1<<13 | v1>>51
	v1 ^= v0
	v0 = v0<<32 | v0>>32

	v2 += v3
	v3 = v3<<16 | v3>>48
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>43
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>47
	v1 ^= v2
	v2 = v2<<32 | v2>>32

	// round 4
	v0 += v1
	v1 = v1<<13 | v1>>51
	v1 ^= v0
	v0 = v0<<32 | v0>>32

	v2 += v3
	v3 = v3<<16 | v3>>48
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>43
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>47
	v1 ^= v2
	v2 = v2<<32 | v2>>32

	hash := v0 ^ v1 ^ v2 ^ v3
	h = hash >> 1
	l = hash << 1 >> 1

	return
}
