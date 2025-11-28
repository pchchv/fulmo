package fulmo

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/pchchv/fulmo/helpers"
)

const (
	itemSize = int64(unsafe.Sizeof(storeItem[any]{}))

	itemNew itemFlag = iota
	itemDelete
	itemUpdate

	hit = iota // keep track of hits and misses
	miss
	keyAdd // keep track of number of keys added, updated and evicted
	keyUpdate
	keyEvict
	costAdd // keep track of cost of keys added and evicted
	costEvict
	dropSets // keep track of how many sets were dropped or rejected later
	rejectSets
	dropGets // keep track of how many gets were kept and dropped on the floor
	keepGets
	doNotUse // should be the final enum. Other enums should be set before this
)

// Item is a full representation of what's stored in the cache for each key-value pair.
type Item[V any] struct {
	flag       itemFlag
	Key        uint64
	Conflict   uint64
	Value      V
	Cost       int64
	Expiration time.Time
	wait       chan struct{}
}

// Key is the generic type to represent the keys type in key-value pair of the cache.
type Key = helpers.Key

type itemFlag byte

type metricType int

// Config is passed to NewCache for creating new Cache instances.
type Config[K Key, V any] struct {
	// NumCounters determines the number of counters (keys) to keep that hold
	// access frequency information. It's generally a good idea to have more
	// counters than the max cache capacity, as this will improve eviction
	// accuracy and subsequent hit ratios.
	//
	// For example, if you expect your cache to hold 1,000,000 items when full,
	// NumCounters should be 10,000,000 (10x). Each counter takes up roughly
	// 3 bytes (4 bits for each counter * 4 copies plus about a byte per
	// counter for the bloom filter). Note that the number of counters is
	// internally rounded up to the nearest power of 2, so the space usage
	// may be a little larger than 3 bytes * NumCounters.
	//
	// Seen good performance in setting this to 10x the number of items
	// you expect to keep in the cache when full.
	NumCounters int64
	// MaxCost is how eviction decisions are made. For example, if MaxCost is
	// 100 and a new item with a cost of 1 increases total cache cost to 101,
	// 1 item will be evicted.
	//
	// MaxCost can be considered as the cache capacity, in whatever units you
	// choose to use.
	//
	// For example, if you want the cache to have a max capacity of 100MB, you
	// would set MaxCost to 100,000,000 and pass an item's number of bytes as
	// the `cost` parameter for calls to Set. If new items are accepted, the
	// eviction process will take care of making room for the new item and not
	// overflowing the MaxCost value.
	//
	// MaxCost could be anything as long as it matches how you're using the cost
	// values when calling Set.
	MaxCost int64
	// BufferItems determines the size of Get buffers.
	//
	// Unless you have a rare use case, using `64` as the BufferItems value
	// results in good performance.
	//
	// If for some reason you see Get performance decreasing with lots of
	// contention (you shouldn't), try increasing this value in increments of 64.
	// This is a fine-tuning mechanism and you probably won't have to touch this.
	BufferItems int64
	// Metrics is true when you want variety of stats about the cache.
	// There is some overhead to keeping statistics, so you should only set this
	// flag to true when testing or throughput performance isn't a major factor.
	Metrics bool
	// OnEvict is called for every eviction with the evicted item.
	OnEvict func(item *Item[V])
	// OnReject is called for every rejection done via the policy.
	OnReject func(item *Item[V])
	// OnExit is called whenever a value is removed from cache. This can be
	// used to do manual memory deallocation. Would also be called on eviction
	// as well as on rejection of the value.
	OnExit func(val V)
	// ShouldUpdate is called when a value already exists in cache and is being updated.
	// If ShouldUpdate returns true, the cache continues with the update (Set). If the
	// function returns false, no changes are made in the cache. If the value doesn't
	// already exist, the cache continue with setting that value for the given key.
	//
	// In this function, you can check whether the new value is valid. For example, if
	// your value has timestamp assosicated with it, you could check whether the new
	// value has the latest timestamp, preventing you from setting an older value.
	ShouldUpdate func(cur, prev V) bool
	// KeyToHash function is used to customize the key hashing algorithm.
	// Each key will be hashed using the provided function. If keyToHash value
	// is not set, the default keyToHash function is used.
	//
	// Ristretto has a variety of defaults depending on the underlying interface type
	// https://github.com/hypermodeinc/ristretto/blob/main/z/z.go#L19-L41).
	//
	// Note that if you want 128bit hashes you should use the both the values
	// in the return of the function. If you want to use 64bit hashes, you can
	// just return the first uint64 and return 0 for the second uint64.
	KeyToHash func(key K) (uint64, uint64)
	// Cost evaluates a value and outputs a corresponding cost. This function is ran
	// after Set is called for a new item or an item is updated with a cost param of 0.
	//
	// Cost is an optional function you can pass to the Config in order to evaluate
	// item cost at runtime, and only whentthe Set call isn't going to be dropped. This
	// is useful if calculating item cost is particularly expensive and you don't want to
	// waste time on items that will be dropped anyways.
	//
	// To signal to Ristretto that you'd like to use this Cost function:
	//   1. Set the Cost field to a non-nil function.
	//   2. When calling Set for new items or item updates, use a `cost` of 0.
	Cost func(value V) int64
	// IgnoreInternalCost set to true indicates to the cache that the cost of
	// internally storing the value should be ignored. This is useful when the
	// cost passed to set is not using bytes as units. Keep in mind that setting
	// this to true will increase the memory usage.
	IgnoreInternalCost bool
	// TtlTickerDurationInSec sets the value of time ticker for cleanup keys on TTL expiry.
	TtlTickerDurationInSec int64
}

// Metrics is a snapshot of performance statistics for the lifetime of a cache instance.
type Metrics struct {
	mu   sync.RWMutex
	all  [doNotUse][]*uint64
	life *helpers.HistogramData // tracks the life expectancy of a key
}

func newMetrics() (s *Metrics) {
	s = &Metrics{
		life: helpers.NewHistogramData(helpers.HistogramBounds(1, 16)),
	}
	for i := 0; i < doNotUse; i++ {
		s.all[i] = make([]*uint64, 256)
		slice := s.all[i]
		for j := range slice {
			slice[j] = new(uint64)
		}
	}
	return
}

// Hits is the number of Get calls where a value was found for the corresponding key.
func (p *Metrics) Hits() uint64 {
	return p.get(hit)
}

// Misses is the number of Get calls where a value was not found for the corresponding key.
func (p *Metrics) Misses() uint64 {
	return p.get(miss)
}

// KeysAdded is the total number of Set calls where a new key-value item was added.
func (p *Metrics) KeysAdded() uint64 {
	return p.get(keyAdd)
}

// KeysUpdated is the total number of Set calls where the value was updated.
func (p *Metrics) KeysUpdated() uint64 {
	return p.get(keyUpdate)
}

// KeysEvicted is the total number of keys evicted.
func (p *Metrics) KeysEvicted() uint64 {
	return p.get(keyEvict)
}

// SetsDropped is the number of Set calls that
// don't make it into internal buffers
// (due to contention or some other reason).
func (p *Metrics) SetsDropped() uint64 {
	return p.get(dropSets)
}

// SetsRejected is the number of Set calls rejected by the policy (TinyLFU).
func (p *Metrics) SetsRejected() uint64 {
	return p.get(rejectSets)
}

// GetsDropped is the number of Get counter increments that are dropped internally.
func (p *Metrics) GetsDropped() uint64 {
	return p.get(dropGets)
}

// GetsKept is the number of Get counter increments that are kept.
func (p *Metrics) GetsKept() uint64 {
	return p.get(keepGets)
}

// CostAdded is the sum of costs that have been added
// (successful Set calls).
func (p *Metrics) CostAdded() uint64 {
	return p.get(costAdd)
}

// CostEvicted is the sum of all costs that have been evicted.
func (p *Metrics) CostEvicted() uint64 {
	return p.get(costEvict)
}

// Ratio is the number of Hits over all accesses (Hits + Misses).
// This is the percentage of successful Get calls.
func (p *Metrics) Ratio() float64 {
	if p == nil {
		return 0.0
	}

	hits, misses := p.get(hit), p.get(miss)
	if hits == 0 && misses == 0 {
		return 0.0
	}

	return float64(hits) / float64(hits+misses)
}

// String returns a string representation of the metrics.
func (p *Metrics) String() string {
	if p == nil {
		return ""
	}

	var buf bytes.Buffer
	for i := 0; i < doNotUse; i++ {
		t := metricType(i)
		fmt.Fprintf(&buf, "%s: %d ", stringFor(t), p.get(t))
	}

	fmt.Fprintf(&buf, "gets-total: %d ", p.get(hit)+p.get(miss))
	fmt.Fprintf(&buf, "hit-ratio: %.2f", p.Ratio())
	return buf.String()
}

// Clear resets all the metrics.
func (p *Metrics) Clear() {
	if p == nil {
		return
	}

	for i := 0; i < doNotUse; i++ {
		for j := range p.all[i] {
			atomic.StoreUint64(p.all[i][j], 0)
		}
	}

	p.mu.Lock()
	p.life = helpers.NewHistogramData(helpers.HistogramBounds(1, 16))
	p.mu.Unlock()
}

func (p *Metrics) LifeExpectancySeconds() *helpers.HistogramData {
	if p == nil {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.life.Copy()
}

func (p *Metrics) add(t metricType, hash, delta uint64) {
	if p == nil {
		return
	}

	valp := p.all[t]
	// avoid false sharing by padding at least 64 bytes of
	// space between two atomic counters which would be incremented
	idx := (hash % 25) * 10
	atomic.AddUint64(valp[idx], delta)
}

func (p *Metrics) get(t metricType) uint64 {
	if p == nil {
		return 0
	}

	var total uint64
	valp := p.all[t]
	for i := range valp {
		total += atomic.LoadUint64(valp[i])
	}

	return total
}

func (p *Metrics) trackEviction(numSeconds int64) {
	if p == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.life.Update(numSeconds)
}

// Cache is a thread-safe hash map implementation with a TinyLFU admission policy and a SampledLFU eviction policy.
// A single Cache instance can be used by any number of goroutines.
type Cache[K Key, V any] struct {
	// storedItems is the central concurrent hashmap where key-value items are stored.
	storedItems store[V]
	// cachePolicy determines what gets let in to the cache and what gets kicked out.
	cachePolicy *defaultPolicy[V]
	// getBuf is a custom ring buffer implementation that gets pushed to when
	// keys are read.
	getBuf *ringBuffer
	// setBuf is a buffer allowing us to batch/drop Sets during times of high
	// contention.
	setBuf chan *Item[V]
	// onEvict is called for item evictions.
	onEvict func(*Item[V])
	// onReject is called when an item is rejected via admission policy.
	onReject func(*Item[V])
	// onExit is called whenever a value goes out of scope from the cache.
	onExit (func(V))
	// KeyToHash function is used to customize the key hashing algorithm.
	// Each key will be hashed using the provided function. If keyToHash value
	// is not set, the default keyToHash function is used.
	keyToHash func(K) (uint64, uint64)
	// stop is used to stop the processItems goroutine.
	stop chan struct{}
	done chan struct{}
	// indicates whether cache is closed.
	isClosed atomic.Bool
	// cost calculates cost from a value.
	cost func(value V) int64
	// ignoreInternalCost dictates whether to ignore the cost of internally storing
	// the item in the cost calculation.
	ignoreInternalCost bool
	// cleanupTicker is used to periodically check for entries whose TTL has passed.
	cleanupTicker *time.Ticker
	// Metrics contains a running log of important statistics like hits, misses,
	// and dropped items.
	Metrics *Metrics
}

// Get returns the value (if any) and
// a boolean representing whether the value was found or not.
// The value can be nil and the boolean can be true at the same time.
// Get will not return expired items.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c == nil || c.isClosed.Load() {
		return zeroValue[V](), false
	}
	keyHash, conflictHash := c.keyToHash(key)

	c.getBuf.Push(keyHash)
	value, ok := c.storedItems.Get(keyHash, conflictHash)
	if ok {
		c.Metrics.add(hit, keyHash, 1)
	} else {
		c.Metrics.add(miss, keyHash, 1)
	}

	return value, ok
}

// processItems is ran by goroutines processing the Set buffer.
func (c *Cache[K, V]) processItems() {
	startTs := make(map[uint64]time.Time)
	numToKeep := 100000 // TODO: Make this configurable via options.
	trackAdmission := func(key uint64) {
		if c.Metrics == nil {
			return
		}

		startTs[key] = time.Now()
		if len(startTs) > numToKeep {
			for k := range startTs {
				if len(startTs) <= numToKeep {
					break
				}
				delete(startTs, k)
			}
		}
	}
	onEvict := func(i *Item[V]) {
		if ts, has := startTs[i.Key]; has {
			c.Metrics.trackEviction(int64(time.Since(ts) / time.Second))
			delete(startTs, i.Key)
		}
		if c.onEvict != nil {
			c.onEvict(i)
		}
	}

	for {
		select {
		case i := <-c.setBuf:
			if i.wait != nil {
				close(i.wait)
				continue
			}
			// calculate item cost value if new or update
			if i.Cost == 0 && c.cost != nil && i.flag != itemDelete {
				i.Cost = c.cost(i.Value)
			}
			if !c.ignoreInternalCost {
				// add the cost of internally storing the object
				i.Cost += itemSize
			}

			switch i.flag {
			case itemNew:
				victims, added := c.cachePolicy.Add(i.Key, i.Cost)
				if added {
					c.storedItems.Set(i)
					c.Metrics.add(keyAdd, i.Key, 1)
					trackAdmission(i.Key)
				} else {
					c.onReject(i)
				}
				for _, victim := range victims {
					victim.Conflict, victim.Value = c.storedItems.Del(victim.Key, 0)
					onEvict(victim)
				}
			case itemUpdate:
				c.cachePolicy.Update(i.Key, i.Cost)
			case itemDelete:
				c.cachePolicy.Del(i.Key) // Deals with metrics updates.
				_, val := c.storedItems.Del(i.Key, i.Conflict)
				c.onExit(val)
			}
		case <-c.cleanupTicker.C:
			c.storedItems.Cleanup(c.cachePolicy, onEvict)
		case <-c.stop:
			c.done <- struct{}{}
			return
		}
	}
}

// collectMetrics just creates a new *Metrics instance and
// adds the pointers to the cache and policy instances.
func (c *Cache[K, V]) collectMetrics() {
	c.Metrics = newMetrics()
	c.cachePolicy.CollectMetrics(c.Metrics)
}

func stringFor(t metricType) string {
	switch t {
	case hit:
		return "hit"
	case miss:
		return "miss"
	case keyAdd:
		return "keys-added"
	case keyUpdate:
		return "keys-updated"
	case keyEvict:
		return "keys-evicted"
	case costAdd:
		return "cost-added"
	case costEvict:
		return "cost-evicted"
	case dropSets:
		return "sets-dropped"
	case rejectSets:
		return "sets-rejected" // by policy.
	case dropGets:
		return "gets-dropped"
	case keepGets:
		return "gets-kept"
	default:
		return "unidentified"
	}
}

func zeroValue[T any]() T {
	var zero T
	return zero
}
