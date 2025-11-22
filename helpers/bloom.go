package helpers

import (
	"log"
	"math"
)

// Bloom filter.
type Bloom struct {
	bitset  []uint64
	ElemNum uint64
	sizeExp uint64
	size    uint64
	setLocs uint64
	shift   uint64
}

// NewBloomFilter returns a new bloomfilter.
func NewBloomFilter(params ...float64) (bloomfilter *Bloom) {
	var entries, locs uint64
	if len(params) == 2 {
		if params[1] < 1 {
			entries, locs = calcSizeByWrongPositives(params[0], params[1])
		} else {
			entries, locs = uint64(params[0]), uint64(params[1])
		}
	} else {
		log.Fatal("usage: New(float64(number_of_entries), float64(number_of_hashlocations))" +
			" i.e. New(float64(1000), float64(3)) or New(float64(number_of_entries)," +
			" float64(number_of_hashlocations)) i.e. New(float64(1000), float64(0.03))")
	}

	size, exponent := getSize(entries)
	bloomfilter = &Bloom{
		sizeExp: exponent,
		size:    size - 1,
		setLocs: locs,
		shift:   64 - exponent,
	}
	bloomfilter.Size(size)
	return bloomfilter
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

func calcSizeByWrongPositives(numEntries, wrongs float64) (uint64, uint64) {
	size := -1 * numEntries * math.Log(wrongs) / math.Pow(float64(0.69314718056), 2)
	return uint64(size), uint64(math.Ceil(float64(0.69314718056) * size / numEntries))
}

func getSize(ui64 uint64) (size uint64, exponent uint64) {
	if ui64 < uint64(512) {
		ui64 = uint64(512)
	}

	size = uint64(1)
	for size < ui64 {
		size <<= 1
		exponent++
	}

	return
}
