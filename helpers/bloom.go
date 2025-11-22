package helpers

// Bloom filter.
type Bloom struct {
	bitset  []uint64
	ElemNum uint64
	sizeExp uint64
	size    uint64
	setLocs uint64
	shift   uint64
}

// Size makes Bloom filter with as bitset of size sz.
func (bl *Bloom) Size(sz uint64) {
	bl.bitset = make([]uint64, sz>>6)
}

// TotalSize returns the total size of the bloom filter.
func (bl *Bloom) TotalSize() int {
	// bl struct has 5 members and each one is 8 byte
	// bitset is a uint64 byte slice
	return len(bl.bitset)*8 + 5*8
}
