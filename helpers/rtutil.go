package helpers

import "unsafe"

type stringStruct struct {
	str unsafe.Pointer
	len int
}

// MemHash is the hash function used by go map,
// it utilizes available hardware instructions(behaves as aeshash if aes instruction is available).
// NOTE: The hash seed changes for every process.
// So, this cannot be used as a persistent hash.
func MemHash(data []byte) uint64 {
	ss := (*stringStruct)(unsafe.Pointer(&data))
	return uint64(memhash(ss.str, 0, uintptr(ss.len)))
}

// MemHashString is the hash function used by go map,
// it utilizes available hardware instructions
// (behaves as aeshash if aes instruction is available).
// NOTE: The hash seed changes for every process.
// So, this cannot be used as a persistent hash.
func MemHashString(str string) uint64 {
	ss := (*stringStruct)(unsafe.Pointer(&str))
	return uint64(memhash(ss.str, 0, uintptr(ss.len)))
}

func Memclr(b []byte) {
	if len(b) == 0 {
		return
	}

	p := unsafe.Pointer(&b[0])
	memclrNoHeapPointers(p, uintptr(len(b)))
}

//go:noescape
//go:linkname memhash runtime.memhash
func memhash(p unsafe.Pointer, h, s uintptr) uintptr

//go:linkname memclrNoHeapPointers runtime.memclrNoHeapPointers
func memclrNoHeapPointers(p unsafe.Pointer, n uintptr)
