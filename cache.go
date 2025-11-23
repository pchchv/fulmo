package fulmo

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/pchchv/fulmo/helpers"
)

const (
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
