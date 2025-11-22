package fulmo

import "fmt"

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
