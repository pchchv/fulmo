package fulmo

import "time"

// TODO: find the optimal value or make it configurable.
var bucketDurationSecs int64 = 5

// bucket type is a map of key to conflict.
type bucket map[uint64]uint64

func storageBucket(t time.Time) int64 {
	return (t.Unix() / bucketDurationSecs) + 1
}

func cleanupBucket(t time.Time) int64 {
	// the bucket to cleanup is always behind the
	// storage bucket by one so that no elements in that bucket
	// (which might not have expired yet) are deleted
	return storageBucket(t) - 1
}
