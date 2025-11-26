package fulmo

type updateFn[V any] func(cur, prev V) bool
