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
