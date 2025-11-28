package fulmo

type testConsumer struct {
	push func([]uint64)
	save bool
}

func (c *testConsumer) Push(items []uint64) bool {
	if c.save {
		c.push(items)
		return true
	}
	return false
}
