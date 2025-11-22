package fulmo

import "testing"

func TestNext2Power(t *testing.T) {
	sz := 12 << 30
	szf := float64(sz) * 0.01
	val := int64(szf)
	t.Logf("szf = %.2f val = %d\n", szf, val)

	pow := next2Power(val)
	t.Logf("pow = %d. mult 4 = %d\n", pow, pow*4)
}

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
