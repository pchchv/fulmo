package helpers

import "unsafe"

var dallocs map[unsafe.Pointer]*dalloc

type dalloc struct {
	t  string
	sz int
}
