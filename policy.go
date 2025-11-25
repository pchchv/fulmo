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

// sampledLFU is an eviction helper storing key-cost pairs.
type sampledLFU struct {
	// NOTE: align maxCost to 64-bit boundary for use with atomic.
	// As per https://golang.org/pkg/sync/atomic/:
	// "On ARM, x86-32, and 32-bit MIPS,
	// it is the caller’s responsibility to arrange
	// for 64-bit alignment of 64-bit words accessed atomically.
	// The first word in a variable or in an allocated struct,
	// array, or slice can be relied upon to be 64-bit aligned."
	used     int64
	maxCost  int64
	metrics  *Metrics
	keyCosts map[uint64]int64
}

func newSampledLFU(maxCost int64) *sampledLFU {
	return &sampledLFU{
		maxCost:  maxCost,
		keyCosts: make(map[uint64]int64),
	}
}

func (p *sampledLFU) add(key uint64, cost int64) {
	p.keyCosts[key] = cost
	p.used += cost
}

func (p *sampledLFU) clear() {
	p.used = 0
	p.keyCosts = make(map[uint64]int64)
}

func (p *sampledLFU) del(key uint64) {
	if cost, ok := p.keyCosts[key]; ok {
		p.used -= cost
		delete(p.keyCosts, key)
		p.metrics.add(costEvict, key, uint64(cost))
		p.metrics.add(keyEvict, key, 1)
	}
}
