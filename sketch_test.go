package fulmo

import "testing"

func BenchmarkSketchIncrement(b *testing.B) {
	b.SetBytes(1)
	s := newCmSketch(16)
	for n := 0; n < b.N; n++ {
		s.Increment(1)
	}
}

func BenchmarkSketchEstimate(b *testing.B) {
	s := newCmSketch(16)
	s.Increment(1)
	b.SetBytes(1)
	for n := 0; n < b.N; n++ {
		s.Estimate(1)
	}
}
