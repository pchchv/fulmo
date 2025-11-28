package fulmo

import "sync"

// ringConsumer is the user-defined object responsible for
// receiving and processing items in batches when buffers are drained.
type ringConsumer interface {
	Push([]uint64) bool
}

// ringBuffer stores multiple buffers (stripes)
// and distributes Pushed items between them to lower contention.
//
// This implements the "batching" process described in
// the BP-Wrapper paper (section III part A).
type ringBuffer struct {
	pool *sync.Pool
}
