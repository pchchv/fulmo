//go:build !jemalloc || !cgo
// +build !jemalloc !cgo

package helpers

import "fmt"

// Provides versions of Calloc, CallocNoRef, etc when jemalloc is not available
// (eg: build without jemalloc tag).

// Free does not do anything in this mode.
func Free(b []byte) {}

// Calloc allocates a slice of size n.
func Calloc(n int, tag string) []byte {
	return make([]byte, n)
}

// CallocNoRef will not give you memory back without jemalloc.
func CallocNoRef(n int, tag string) []byte {
	// We do the add here just to stay compatible with a corresponding Free call.
	return nil
}

func StatsPrint() {
	fmt.Println("Using Go memory")
}

// ReadMemStats doesn't do anything since all the memory is
// being managed by the Go runtime.
func ReadMemStats(_ *MemStats) {}

func Leaks() string {
	return "Leaks: Using Go memory"
}
