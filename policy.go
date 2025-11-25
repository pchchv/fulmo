package fulmo

import (
	"sync"
	"sync/atomic"

	"github.com/pchchv/fulmo/helpers"
)

// lfuSample is the number of items to sample when looking at eviction candidates.
// 5 seems to be the most optimal number [citation needed].
const lfuSample = 5

type policyPair struct {
	key  uint64
	cost int64
}

func newPolicy[V any](numCounters, maxCost int64) *defaultPolicy[V] {
	return newDefaultPolicy[V](numCounters, maxCost)
}

type defaultPolicy[V any] struct {
	sync.Mutex
	isClosed bool
	metrics  *Metrics
	admit    *tinyLFU
	evict    *sampledLFU
	stop     chan struct{}
	done     chan struct{}
	itemsCh  chan []uint64
}

func newDefaultPolicy[V any](numCounters, maxCost int64) *defaultPolicy[V] {
	p := &defaultPolicy[V]{
		admit:   newTinyLFU(numCounters),
		evict:   newSampledLFU(maxCost),
		itemsCh: make(chan []uint64, 3),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}

	go p.processItems()
	return p
}

func (p *defaultPolicy[V]) Close() {
	if p.isClosed {
		return
	}

	// block until the p.processItems goroutine returns
	p.stop <- struct{}{}
	<-p.done
	close(p.stop)
	close(p.done)
	close(p.itemsCh)
	p.isClosed = true
}

func (p *defaultPolicy[V]) Clear() {
	p.Lock()
	p.admit.clear()
	p.evict.clear()
	p.Unlock()
}

func (p *defaultPolicy[V]) Update(key uint64, cost int64) {
	p.Lock()
	p.evict.updateIfHas(key, cost)
	p.Unlock()
}

func (p *defaultPolicy[V]) processItems() {
	for {
		select {
		case items := <-p.itemsCh:
			p.Lock()
			p.admit.Push(items)
			p.Unlock()
		case <-p.stop:
			p.done <- struct{}{}
			return
		}
	}
}

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

func (p *sampledLFU) getMaxCost() int64 {
	return atomic.LoadInt64(&p.maxCost)
}

func (p *sampledLFU) updateMaxCost(maxCost int64) {
	atomic.StoreInt64(&p.maxCost, maxCost)
}

func (p *sampledLFU) updateIfHas(key uint64, cost int64) bool {
	if prev, found := p.keyCosts[key]; found {
		// update the cost of an existing key,
		// but don't worry about evicting
		// evictions will be handled the next time a new item is added
		p.metrics.add(keyUpdate, key, 1)
		if prev > cost {
			diff := prev - cost
			p.metrics.add(costAdd, key, ^(uint64(diff) - 1))
		} else if cost > prev {
			diff := cost - prev
			p.metrics.add(costAdd, key, uint64(diff))
		}

		p.used += cost - prev
		p.keyCosts[key] = cost
		return true
	}
	return false
}

func (p *sampledLFU) roomLeft(cost int64) int64 {
	return p.getMaxCost() - (p.used + cost)
}

func (p *sampledLFU) fillSample(in []*policyPair) []*policyPair {
	if len(in) >= lfuSample {
		return in
	}

	for key, cost := range p.keyCosts {
		in = append(in, &policyPair{key, cost})
		if len(in) >= lfuSample {
			return in
		}
	}

	return in
}
