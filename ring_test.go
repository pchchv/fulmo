package fulmo

type testConsumer struct {
	push func([]uint64)
	save bool
}
