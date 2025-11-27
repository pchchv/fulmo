package fulmo

import "time"

type updateFn[V any] func(cur, prev V) bool

type storeItem[V any] struct {
	key        uint64
	value      V
	conflict   uint64
	expiration time.Time
}
