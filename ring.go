package fulmo

// ringConsumer is the user-defined object responsible for
// receiving and processing items in batches when buffers are drained.
type ringConsumer interface {
	Push([]uint64) bool
}
