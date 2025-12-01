package fulmo

// Clairvoyant is a mock cache providin optimal hit ratios to the Fulmo.
// It looks ahead and evicts the absolute least valuable item that
// comes closest to the actual cache.
type Clairvoyant struct {
	capacity uint64
	access   []uint64
	hits     map[uint64]uint64
}

func NewClairvoyant(capacity uint64) *Clairvoyant {
	return &Clairvoyant{
		capacity: capacity,
		hits:     make(map[uint64]uint64),
		access:   make([]uint64, 0),
	}
}

// Set isn't important because it is only called after a Get
// (in the case of hit ratio benchmarks, at least).
func (c *Clairvoyant) Set(key, value interface{}, cost int64) bool {
	return false
}

// Get just records the cache access so that it can later take
// this event into consideration when calculating the absolute least valuable item to evict.
func (c *Clairvoyant) Get(key interface{}) (interface{}, bool) {
	c.hits[key.(uint64)]++
	c.access = append(c.access, key.(uint64))
	return nil, false
}

type clairvoyantItem struct {
	key  uint64
	hits uint64
}
