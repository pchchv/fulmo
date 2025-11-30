package helpers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalloc(t *testing.T) {
	// Checking if jemalloc is being used.
	// JE_MALLOC_CONF="abort:true,tcache:false"

	StatsPrint()
	buf := CallocNoRef(1, "test")
	if len(buf) == 0 {
		t.Skipf("Not using jemalloc. Skipping test.")
	}
	Free(buf)
	require.Equal(t, int64(0), NumAllocBytes())

	buf1 := Calloc(128, "test")
	require.Equal(t, int64(128), NumAllocBytes())
	buf2 := Calloc(128, "test")
	require.Equal(t, int64(256), NumAllocBytes())

	Free(buf1)
	require.Equal(t, int64(128), NumAllocBytes())

	// _ = buf2
	Free(buf2)
	require.Equal(t, int64(0), NumAllocBytes())
	fmt.Println(Leaks())

	// Double free would panic when debug mode is enabled in jemalloc.
	// Free(buf2)
	// require.Equal(t, int64(0), NumAllocBytes())
}
