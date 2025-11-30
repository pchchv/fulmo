package sim

import (
	"bufio"
	"io"
	"math/rand"
	"time"
)

// Parser is used as a parameter to NewReader,
// allowing easy creation of Simulators from various trace file formats.
type Parser func(string, error) ([]uint64, error)

// Simulator is the core type of the `sim` package.
// It is a function that returns a key from some source
// (composed of other functions in this package, generated or parsed).
// Simulators can be used to approximate access distributions.
type Simulator func() (uint64, error)

// NewReader creates a Simulator from two components:
// the Parser, which is a filetype specific function for parsing lines,
// and the file itself, which will be read from.
//
// When every line in the file has been read, ErrDone will be returned.
// For some trace formats (LIRS) there is one item per line.
// For others (ARC) there is a range of items on each line.
// Thus, the true number of items in each file is hard to determine,
// so it's up to the user to handle ErrDone accordingly.
func NewReader(parser Parser, file io.Reader) Simulator {
	var err error
	i := -1
	s := make([]uint64, 0)
	b := bufio.NewReader(file)
	return func() (uint64, error) {
		// only parse a new line when we've run out of items
		if i++; i == len(s) {
			// parse sequence from line
			if s, err = parser(b.ReadString('\n')); err != nil {
				s = []uint64{0}
			}
			i = 0
		}
		return s[i], err
	}
}

// NewZipfian creates a Simulator returning numbers following a
// Zipfian distribution infinitely.
// Zipfian distributions are useful for simulating real workloads.
func NewZipfian(s, v float64, n uint64) Simulator {
	z := rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), s, v, n)
	return func() (uint64, error) {
		return z.Uint64(), nil
	}
}

// NewUniform creates a Simulator returning
// uniformly distributed random numbers [0, max) infinitely.
func NewUniform(max uint64) Simulator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() (uint64, error) {
		return uint64(r.Int63n(int64(max))), nil
	}
}
