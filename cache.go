package fulmo

import (
	"sync"
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
