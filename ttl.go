package fulmo

import "time"

// TODO: find the optimal value or make it configurable.
var bucketDurationSecs int64 = 5

// bucket type is a map of key to conflict.
type bucket map[uint64]uint64

func storageBucket(t time.Time) int64 {
	return (t.Unix() / bucketDurationSecs) + 1
}
