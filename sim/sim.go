package sim

import (
	"bufio"
	"io"
)


// Parser is used as a parameter to NewReader,
// allowing easy creation of Simulators from various trace file formats.
type Parser func(string, error) ([]uint64, error)

// Simulator is the core type of the `sim` package.
// It is a function that returns a key from some source
// (composed of other functions in this package, generated or parsed).
// Simulators can be used to approximate access distributions.
type Simulator func() (uint64, error)
