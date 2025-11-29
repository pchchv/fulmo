package helpers

import "unsafe"

var dallocs map[unsafe.Pointer]*dalloc

type dalloc struct {
	t  string
	sz int
}

// By initializing dallocs, it,s possible to begin tracking memory allocation and deallocation through helpers.Calloc.
func init() {
	dallocs = make(map[unsafe.Pointer]*dalloc)
}
