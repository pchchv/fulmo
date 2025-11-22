package helpers

// Bloom filter
type Bloom struct {
	bitset  []uint64
	ElemNum uint64
	sizeExp uint64
	size    uint64
	setLocs uint64
	shift   uint64
}
