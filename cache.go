package fulmo

import "time"

const (
	itemNew itemFlag = iota
	itemDelete
	itemUpdate
)

// Item is a full representation of what's stored in the cache for each key-value pair.
type Item[V any] struct {
	flag       itemFlag
	Key        uint64
	Conflict   uint64
	Value      V
	Cost       int64
	Expiration time.Time
	wait       chan struct{}
}

type itemFlag byte

type metricType int

