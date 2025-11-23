package fulmo

import "github.com/pchchv/fulmo/helpers"

// tinyLFU is an admission helper that tracks
// access frequency using tiny (4-bit) counters in
// the form of a count-min sketch.
// tinyLFU is NOT thread-safe.
type tinyLFU struct {
	resetAt int64
	incrs   int64
	freq    *cmSketch
	door    *helpers.Bloom
}

func newTinyLFU(numCounters int64) *tinyLFU {
	return &tinyLFU{
		freq:    newCmSketch(numCounters),
		door:    helpers.NewBloomFilter(float64(numCounters), 0.01),
		resetAt: numCounters,
	}
}
