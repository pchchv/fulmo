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

// ringStripe is a singular ring buffer that is not concurrent safe.
type ringStripe struct {
	cons ringConsumer
	data []uint64
	capa int
}

func newRingStripe(cons ringConsumer, capa int64) *ringStripe {
	return &ringStripe{
		cons: cons,
		data: make([]uint64, 0, capa),
		capa: int(capa),
	}
}
