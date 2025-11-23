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

func (p *tinyLFU) Estimate(key uint64) int64 {
	hits := p.freq.Estimate(key)
	if p.door.Has(key) {
		hits++
	}
	return hits
}

func (p *tinyLFU) Increment(key uint64) {
	// flip doorkeeper bit if not already done
	if added := p.door.AddIfNotHas(key); !added {
		// increment count-min counter if doorkeeper bit is already set
		p.freq.Increment(key)
	}

	p.incrs++
	if p.incrs >= p.resetAt {
		p.reset()
	}
}

func (p *tinyLFU) Push(keys []uint64) {
	for _, key := range keys {
		p.Increment(key)
	}
}

func (p *tinyLFU) clear() {
	p.incrs = 0
	p.door.Clear()
	p.freq.Clear()
}

func (p *tinyLFU) reset() {
	// zero out incrs
	p.incrs = 0
	// clears doorkeeper bits
	p.door.Clear()
	// halves count-min counters
	p.freq.Reset()
}
