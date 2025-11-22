package fulmo

import (
	"fmt"
	"math/rand"
	"time"
)

// cmDepth is the number of counter copies to store (think of it as rows).
const cmDepth = 4

// cmRow is a row of bytes, with each byte holding two counters.
type cmRow []byte

func newCmRow(numCounters int64) cmRow {
	return make(cmRow, numCounters/2)
}

func (r cmRow) get(n uint64) byte {
	return (r[n/2] >> ((n & 1) * 4)) & 0x0f
}

func (r cmRow) clear() {
	// zero each counter
	for i := range r {
		r[i] = 0
	}
}

func (r cmRow) reset() {
	// halve each counter
	for i := range r {
		r[i] = (r[i] >> 1) & 0x77
	}
}

func (r cmRow) string() (s string) {
	for i := uint64(0); i < uint64(len(r)*2); i++ {
		s += fmt.Sprintf("%02d ", (r[(i/2)]>>((i&1)*4))&0x0f)
	}
	return s[:len(s)-1]
}

func (r cmRow) increment(n uint64) {
	// index of the counter
	i := n / 2
	// shift distance (even 0, odd 4)
	s := (n & 1) * 4
	// counter value
	v := (r[i] >> s) & 0x0f
	// only increment if not max value (overflow wrap is bad for LFU)
	if v < 15 {
		r[i] += 1 << s
	}
}

// cmSketch is a Count-Min sketch implementation with 4-bit counters,
// heavily based on Damian Gryski's CM4.
type cmSketch struct {
	mask uint64
	rows [cmDepth]cmRow
	seed [cmDepth]uint64
}

func newCmSketch(numCounters int64) *cmSketch {
	if numCounters == 0 {
		panic("cmSketch: bad numCounters")
	}

	// get the next power of 2 for better cache performance
	numCounters = next2Power(numCounters)
	sketch := &cmSketch{mask: uint64(numCounters - 1)}
	// initialize rows of counters and seeds
	// cryptographic precision not needed
	source := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	for i := 0; i < cmDepth; i++ {
		sketch.seed[i] = source.Uint64()
		sketch.rows[i] = newCmRow(numCounters)
	}

	return sketch
}

// Clear zeroes all counters.
func (s *cmSketch) Clear() {
	for _, r := range s.rows {
		r.clear()
	}
}

// Reset halves all counter values.
func (s *cmSketch) Reset() {
	for _, r := range s.rows {
		r.reset()
	}
}

// next2Power rounds x up to the next power of 2,
// if it's not already one.
func next2Power(x int64) int64 {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	x++
	return x
}
