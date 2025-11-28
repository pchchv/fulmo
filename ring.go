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

// newRingBuffer returns a striped ring buffer.
// The Consumer in ringConfig will be called when individual stripes are
// full and need to drain their elements.
func newRingBuffer(cons ringConsumer, capa int64) *ringBuffer {
	// LOSSY buffers use a very simple sync.Pool for concurrently reusing stripes.
	// Some stripes are lost due to GC (unclaimed items in the sync.Pool are cleared),
	// but the performance gains generally outweigh the small percentage of lost elements.
	// This performance is primarily comes from low-level runtime functions used in
	// the standard library (e.g., runtime_procPin()).
	return &ringBuffer{
		pool: &sync.Pool{
			New: func() interface{} { return newRingStripe(cons, capa) },
		},
	}
}

// Push adds an element to one of
// the internal stripes and possibly drains if
// the stripe becomes full.
func (b *ringBuffer) Push(item uint64) {
	// reuse or create a new stripe
	stripe := b.pool.Get().(*ringStripe)
	stripe.Push(item)
	b.pool.Put(stripe)
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

// Push appends an item in the ring buffer and
// drains (copies items and sends to Consumer) if full.
func (s *ringStripe) Push(item uint64) {
	s.data = append(s.data, item)
	// decide if the ring buffer should be drained
	if len(s.data) >= s.capa {
		// send elements to consumer and create a new ring stripe
		if s.cons.Push(s.data) {
			s.data = make([]uint64, 0, s.capa)
		} else {
			s.data = s.data[:0]
		}
	}
}
