//go:build jemalloc
// +build jemalloc

package helpers

/*
#cgo LDFLAGS: /usr/local/lib/libjemalloc.a -L/usr/local/lib -Wl,-rpath,/usr/local/lib -ljemalloc -lm -lstdc++ -pthread -ldl
#include <stdlib.h>
#include <jemalloc/jemalloc.h>
*/
import "C"
import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	dallocs   map[unsafe.Pointer]*dalloc
	dallocsMu sync.Mutex
)

// The go:linkname directives provides backdoor access to private functions in
// the runtime. Below we're accessing the throw function.

//go:linkname throw runtime.throw
func throw(s string)

// New allocates a slice of size n. The returned slice is from manually managed
// memory and MUST be released by calling Free. Failure to do so will result in
// a memory leak.
//
// Compile jemalloc with ./configure --with-jemalloc-prefix="je_"
// https://android.googlesource.com/platform/external/jemalloc_new/+/6840b22e8e11cb68b493297a5cd757d6eaa0b406/TUNING.md
// These two config options seems useful for frequent allocations and deallocations in
// multi-threaded programs (like we have).
// JE_MALLOC_CONF="background_thread:true,metadata_thp:auto"
//
// Compile Go program with `go build -tags=jemalloc` to enable this.

type dalloc struct {
	t  string
	sz int
}

// Free frees the specified slice.
func Free(b []byte) {
	if sz := cap(b); sz != 0 {
		b = b[:cap(b)]
		ptr := unsafe.Pointer(&b[0])
		C.je_free(ptr)
		atomic.AddInt64(&numBytes, -int64(sz))
		dallocsMu.Lock()
		delete(dallocs, ptr)
		dallocsMu.Unlock()
	}
}

// By initializing dallocs, it,s possible to begin tracking memory allocation and deallocation through helpers.Calloc.
func init() {
	dallocs = make(map[unsafe.Pointer]*dalloc)
}
